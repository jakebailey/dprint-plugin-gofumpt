import binaryen from "binaryen";
import { $ as _$ } from "execa";
import { task } from "hereby";
import assert from "node:assert";
import fs from "node:fs";
import path from "node:path";
import { parseArgs } from "node:util";
import tmp from "tmp";

const $ = _$({ verbose: "short", stdio: "inherit" });
const $pipe = _$({ verbose: "short" });

const { values: options } = parseArgs({
    args: process.argv.slice(2),
    options: {
        docker: { type: "boolean" },
    },
    strict: false,
    allowPositionals: true,
});

const metadataDir = "metadata";

async function writeIfChanged(filePath: string, content: string) {
    try {
        const existing = await fs.promises.readFile(filePath, "utf8");
        if (existing === content) {
            console.log(`Skipped ${filePath} (unchanged)`);
            return;
        }
    } catch {
        // File doesn't exist, will write
    }
    await fs.promises.writeFile(filePath, content);
    console.log(`Generated ${filePath}`);
}

async function generateVersion() {
    const packageJson = JSON.parse(await fs.promises.readFile("package.json", "utf8"));
    const version: string = packageJson.version;
    const versionFile = path.join(metadataDir, "VERSION");
    await writeIfChanged(versionFile, version);
    return version;
}

async function generateSchema(version: string) {
    const schema = {
        $schema: "http://json-schema.org/draft-07/schema#",
        $id: `https://plugins.dprint.dev/jakebailey/gofumpt/v${version}/schema.json`,
        title: "Config",
        description: "Configuration for dprint-plugin-gofumpt. "
            + "It is recommended to set both langVersion and modulePath for consistent formatting results.",
        type: "object",
        properties: {
            langVersion: {
                description: 'The Go language version to target (e.g., "go1.24", "go1.25"). '
                    + 'Must start with "go" prefix. '
                    + 'If empty, defaults to "go1". '
                    + "Recommended to set explicitly.",
                type: "string",
                default: "",
                pattern: "^(go(0|[1-9][0-9]*)(\\.( 0|[1-9][0-9]*))?(\\.( 0|[1-9][0-9]*)|[a-z]+(0|[1-9][0-9]*)?)?)?$",
                examples: ["go1", "go1.24", "go1.25", "go1.24.1", "go1.25rc1", "go1.25beta2"],
            },
            modulePath: {
                description: "The module path of the package being formatted. "
                    + "Used for import sorting. "
                    + "Recommended to set explicitly.",
                type: "string",
                default: "",
                examples: ["github.com/example/myproject"],
            },
            extraRules: {
                description: "Enable extra formatting rules beyond the default gofumpt rules. "
                    + "These rules are stricter but may not be desired in all projects.",
                type: "boolean",
                default: false,
            },
        },
    };
    const schemaFile = path.join(metadataDir, "schema.json");
    await writeIfChanged(schemaFile, JSON.stringify(schema, null, 4) + "\n");
}

async function generateLicenses() {
    const license = await fs.promises.readFile("LICENSE", "utf8");

    const separator = `
================================================================================
THIRD PARTY LICENSES
================================================================================
`;

    const goMod = await fs.promises.readFile("go.mod", "utf8");
    const moduleName = goMod.match(/^module\s+(\S+)/m)?.[1];
    if (!moduleName) {
        throw new Error("Could not find module name in go.mod");
    }

    const template = `{{ range . }}
================================================================================
{{ .Name }}{{ if .Version }} {{ .Version }}{{ end }}
{{ .LicenseName }}
================================================================================

{{ .LicenseText }}
{{ end }}
`;
    const templateFile = tmp.fileSync({ postfix: ".tpl" });
    await fs.promises.writeFile(templateFile.name, template);

    const { stdout } = await $pipe({
        env: { GOFLAGS: "-tags=tinygo" },
    })`go run github.com/google/go-licenses/v2@latest report . --ignore=${moduleName} --template=${templateFile.name}`;

    templateFile.removeCallback();

    const content = (license + separator + stdout).trimEnd() + "\n";
    const licensesFile = path.join(metadataDir, "LICENSES");
    await writeIfChanged(licensesFile, content);
}

export const metadata = task({
    name: "metadata",
    description: "Generates the metadata files.",
    run: async () => {
        await fs.promises.mkdir(metadataDir, { recursive: true });
        const version = await generateVersion();
        await Promise.all([
            generateSchema(version),
            generateLicenses(),
        ]);
    },
});

const WASM_FILE = "plugin.wasm";
const TINYGO_VERSION = "0.40.0";
const DOCKER_IMAGE = `tinygo/tinygo:${TINYGO_VERSION}`;

async function patchWasm(wasmBinary: Uint8Array) {
    const module = binaryen.readBinary(wasmBinary);

    // Set _initialize as the start function
    module.setStart(module.getFunction("_initialize"));

    const output = module.emitBinary();
    await fs.promises.writeFile(WASM_FILE, output);
    module.dispose();
}

const tinygoArgs = [
    "build",
    "-target=wasm-unknown",
    "-scheduler=none",
    "-no-debug",
    "-opt=2",
];

async function runBuild(useDocker: boolean) {
    let wasmBinary: Uint8Array;
    if (useDocker) {
        /* dprint-ignore-start */
        const dockerArgs = [
            "run",
            "--rm",
            "-v", `${process.cwd()}:/src`,
            "-w", "/src",
            "-e", "GOFLAGS=-buildvcs=false",
            DOCKER_IMAGE,
            "tinygo",
            ...tinygoArgs,
            "-o", "/dev/stdout",
        ];
        /* dprint-ignore-end */
        const { stdout } = await $pipe({ encoding: "buffer" })`docker ${dockerArgs}`;
        wasmBinary = stdout;
    } else {
        await $`tinygo ${tinygoArgs} -o ${WASM_FILE}`;
        wasmBinary = await fs.promises.readFile(WASM_FILE);
    }
    await patchWasm(wasmBinary);
}

export const build = task({
    name: "build",
    description: "Builds the WASM plugin. Use --docker to build via Docker.",
    dependencies: [metadata],
    run: async () => {
        await runBuild(!!options.docker);
    },
});

export const test = task({
    name: "test",
    description: "Builds and runs tests.",
    dependencies: [build],
    run: async () => {
        const cacheDir = tmp.dirSync({ unsafeCleanup: true });
        const testDir = path.join("testdata", "basic");
        await fs.promises.copyFile(path.join(testDir, "input.go.txt"), path.join(testDir, "test.go"));
        try {
            await $({
                cwd: testDir,
                env: { DPRINT_CACHE_DIR: cacheDir.name },
            })`dprint fmt --log-level=debug --incremental=false`;
            const expected = await fs.promises.readFile(path.join(testDir, "expected.go"), "utf8");
            const actual = await fs.promises.readFile(path.join(testDir, "test.go"), "utf8");
            assert.strictEqual(actual, expected, "Formatted output does not match expected");
        } finally {
            await fs.promises.rm(path.join(testDir, "test.go"), { force: true });
            cacheDir.removeCallback();
        }
    },
});
