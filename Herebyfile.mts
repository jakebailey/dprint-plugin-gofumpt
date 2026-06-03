import binaryen from "binaryen";
import { $ as _$ } from "execa";
import { task } from "hereby";
import assert from "node:assert";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { parseArgs } from "node:util";

const $ = _$({ verbose: "short", stdio: "inherit" });
const $pipe = _$({ verbose: "short" });
const $quiet = _$({});

const { values: options } = parseArgs({
    args: process.argv.slice(2),
    options: {
        docker: { type: "boolean", default: true },
    },
    strict: false,
    allowPositionals: true,
    allowNegative: true,
});

const metadataDir = "metadata";

async function createTempDir(prefix: string) {
    const dir = await fs.promises.mkdtemp(path.join(os.tmpdir(), prefix));
    return {
        dir,
        cleanup: async () => {
            await fs.promises.rm(dir, { recursive: true, force: true });
        },
    };
}

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
    await writeIfChanged(versionFile, version + "\n");
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
                pattern: "^(go(0|[1-9][0-9]*)(\\.((0|[1-9][0-9]*)))?(\\.((0|[1-9][0-9]*))|[a-z]+(0|[1-9][0-9]*)?)?)?$",
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
                    + "These rules are stricter but may not be desired in all projects. "
                    + "Deprecated: use extra to enable individual rules instead.",
                type: "boolean",
                default: false,
            },
            extra: {
                description: "Enable individual extra formatting rules beyond the default gofumpt rules.",
                type: "object",
                default: {},
                properties: {
                    groupParams: {
                        description: "Group adjacent function parameters with repeated types.",
                        type: "boolean",
                        default: false,
                    },
                    clotheReturns: {
                        description:
                            "Replace naked returns in functions with named results with explicit return values.",
                        type: "boolean",
                        default: false,
                    },
                },
            },
        },
    };
    const schemaFile = path.join(metadataDir, "schema.json");
    await writeIfChanged(schemaFile, JSON.stringify(schema, null, 4) + "\n");
}

async function updateReadmeVersion(version: string) {
    const readmePath = "README.md";
    const readme = await fs.promises.readFile(readmePath, "utf8");
    const updated = readme.replace(
        /gofumpt-v[0-9]+\.[0-9]+\.[0-9]+\.wasm/g,
        `gofumpt-v${version}.wasm`,
    );
    await writeIfChanged(readmePath, updated);
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
    const templateFile = await createTempDir("dprint-plugin-gofumpt-template-");
    const templatePath = path.join(templateFile.dir, "template.tpl");
    await fs.promises.writeFile(templatePath, template);

    const { stdout } = await $pipe({
        env: { GOFLAGS: "-tags=tinygo -mod=mod" },
    })`go run github.com/google/go-licenses/v2@v2.0.1 report . --ignore=${moduleName} --template=${templatePath}`;

    await templateFile.cleanup();

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
            updateReadmeVersion(version),
        ]);
    },
});

const WASM_FILE = "plugin.wasm";
// renovate: datasource=github-releases depName=tinygo-org/tinygo
const TINYGO_VERSION = "0.41.1";
const DOCKER_IMAGE = `ghcr.io/tinygo-org/tinygo:${TINYGO_VERSION}`;

const pullTinygo = task({
    name: "pullTinygo",
    description: "Pulls the TinyGo Docker image if not already present.",
    run: async () => {
        const { stdout } = await $pipe`docker images -q ${DOCKER_IMAGE}`;
        if (stdout.trim()) {
            console.log(`Image ${DOCKER_IMAGE} already present.`);
            return;
        }
        await $`docker pull ${DOCKER_IMAGE}`;
    },
});

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
    description: "Builds the WASM plugin. Use --no-docker to build locally.",
    dependencies: options.docker ? [metadata, pullTinygo] : [metadata],
    run: async () => {
        await runBuild(!!options.docker);
    },
});

async function runTest(testName: string) {
    const cacheDir = await createTempDir("dprint-plugin-gofumpt-cache-");
    const testDir = path.join("testdata", testName);
    await fs.promises.copyFile(path.join(testDir, "input.go.txt"), path.join(testDir, "test.go"));
    try {
        const result = await $quiet({
            cwd: testDir,
            env: { DPRINT_CACHE_DIR: cacheDir.dir },
            all: true,
        })`dprint fmt --log-level=debug --incremental=false`;
        try {
            const expected = await fs.promises.readFile(path.join(testDir, "expected.go"), "utf8");
            const actual = await fs.promises.readFile(path.join(testDir, "test.go"), "utf8");
            assert.strictEqual(actual, expected, `Formatted output does not match expected for test "${testName}"`);
        } catch (e) {
            process.stderr.write(result.all ?? "");
            throw e;
        }
    } finally {
        await fs.promises.rm(path.join(testDir, "test.go"), { force: true });
        await cacheDir.cleanup();
    }
}

// txtar format: https://pkg.go.dev/golang.org/x/tools/txtar
// Comment lines at the top, then files separated by "-- filename --" markers.
interface TxtarFile {
    name: string;
    data: string;
}

interface TxtarArchive {
    comment: string;
    files: TxtarFile[];
}

function parseTxtar(content: string): TxtarArchive {
    const files: TxtarFile[] = [];
    const commentLines: string[] = [];
    let currentFileName: string | null = null;
    let currentDataLines: string[] = [];
    const fileMarker = /^-- (.+) --$/;

    for (const line of content.split("\n")) {
        const match = line.match(fileMarker);
        if (match) {
            if (currentFileName !== null) {
                // The \n before the marker belongs to the previous file's data
                files.push({ name: currentFileName, data: currentDataLines.join("\n") + "\n" });
            }
            currentFileName = match[1];
            currentDataLines = [];
        } else if (currentFileName !== null) {
            currentDataLines.push(line);
        } else {
            commentLines.push(line);
        }
    }
    if (currentFileName !== null) {
        // Last file: trailing \n already represented by empty string in split
        files.push({ name: currentFileName, data: currentDataLines.join("\n") });
    }

    return { comment: commentLines.join("\n"), files };
}

interface GofumptTestCase {
    inputFile: string;
    goldenFile: string;
    extraRules: boolean;
    extra: {
        groupParams: boolean;
        clotheReturns: boolean;
    };
    langVersion: string;
}

function parseExtraRules(args: string[]): Pick<GofumptTestCase, "extraRules" | "extra"> {
    const extraArg = args.find((a) => a === "-extra" || a.startsWith("-extra="));
    const extra = {
        groupParams: false,
        clotheReturns: false,
    };
    if (!extraArg) {
        return { extraRules: false, extra };
    }

    const value = extraArg === "-extra" ? "true" : extraArg.slice("-extra=".length);
    if (value === "true") {
        return { extraRules: true, extra };
    }
    if (value === "false") {
        return { extraRules: false, extra };
    }

    for (const rule of value.split(",")) {
        switch (rule) {
            case "group_params":
                extra.groupParams = true;
                break;
            case "clothe_returns":
                extra.clotheReturns = true;
                break;
            default:
                throw new Error(`Unknown gofumpt extra rule: ${rule}`);
        }
    }

    return { extraRules: false, extra };
}

function parseGofumptTests(archive: TxtarArchive): GofumptTestCase[] {
    const cases: GofumptTestCase[] = [];
    const lines = archive.comment.split("\n");

    // Track go.mod lang version mutations from `exec go mod edit -go=X`
    const goMod = archive.files.find((f) => f.name === "go.mod");
    const goDirective = goMod?.data.match(/^go\s+(\S+)/m);
    let currentGoVersion = goDirective ? goDirective[1] : "";

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();

        // Track: exec go mod edit -go=X
        const goModEditMatch = line.match(/^exec go mod edit -go=(\S+)/);
        if (goModEditMatch) {
            currentGoVersion = goModEditMatch[1];
            continue;
        }

        // Match: exec gofumpt [flags] -w foo.go [bar.go ...]
        // or: exec gofumpt [flags] foo.go (stdout mode, next line: cmp stdout foo.go.golden)
        const execMatch = line.match(/^exec gofumpt\s+(.*)$/);
        if (!execMatch) continue;

        const args = execMatch[1].split(/\s+/);
        const extraRules = parseExtraRules(args);
        const langArg = args.find((a) => a.startsWith("-lang="));
        const langVersion = langArg ? langArg.replace("-lang=", "") : currentGoVersion ? "go" + currentGoVersion : "";

        const isWrite = args.includes("-w");
        const goFiles = args.filter((a) => a.endsWith(".go") && !a.endsWith(".golden"));

        if (goFiles.length === 0) continue;

        // Check for -d or -l only commands (idempotency checks, listing); skip them
        if (args.includes("-d") || args.includes("-l")) continue;

        if (isWrite) {
            // Next line should be: cmp foo.go foo.go.golden
            for (let j = i + 1; j < lines.length && j <= i + goFiles.length; j++) {
                const cmpMatch = lines[j].trim().match(/^cmp\s+(\S+)\s+(\S+)$/);
                if (cmpMatch) {
                    cases.push({
                        inputFile: cmpMatch[1],
                        goldenFile: cmpMatch[2],
                        ...extraRules,
                        langVersion,
                    });
                }
            }
        } else {
            // stdout mode: next line is cmp stdout foo.go.golden
            const nextLine = lines[i + 1]?.trim();
            const cmpMatch = nextLine?.match(/^cmp\s+stdout\s+(\S+)$/);
            if (cmpMatch && goFiles.length === 1) {
                cases.push({
                    inputFile: goFiles[0],
                    goldenFile: cmpMatch[1],
                    ...extraRules,
                    langVersion,
                });
            }
        }
    }

    return cases;
}

async function getGofumptTestdataDir(): Promise<string> {
    const { stdout } = await $pipe({
        env: { GOFLAGS: "-tags=tinygo -mod=mod" },
    })`go list -m -json mvdan.cc/gofumpt`;
    const info = JSON.parse(stdout);
    return path.join(info.Dir, "testdata", "script");
}

// Tests that are too complex or CLI-specific to run through the dprint plugin.
const skippedTxtarTests = new Set([
    "deprecated-flags", // Tests CLI flag deprecation warnings
    "diagnose", // Tests gofumpt:diagnose comments
    "diff", // Tests -d flag behavior
    "gomod", // Tests go.mod edge cases with multiple modules
    "ignore", // Tests vendor/testdata ignore behavior
    "long-lines", // Tests GOFUMPT_SPLIT_LONG_LINES env var
    "workspaces", // Tests go.work behavior
]);

async function runGofumptTxtarTest(txtarPath: string) {
    const content = await fs.promises.readFile(txtarPath, "utf8");
    const archive = parseTxtar(content);
    const testCases = parseGofumptTests(archive);

    if (testCases.length === 0) {
        console.log(`  (no applicable test cases found, skipping)`);
        return;
    }

    const fileMap = new Map(archive.files.map((f) => [f.name, f.data]));

    for (const tc of testCases) {
        const inputData = fileMap.get(tc.inputFile);
        const goldenData = fileMap.get(tc.goldenFile);
        if (!inputData || !goldenData) {
            throw new Error(`Missing file in archive: ${tc.inputFile} or ${tc.goldenFile}`);
        }

        const testDir = await createTempDir("dprint-plugin-gofumpt-test-");
        const cacheDir = await createTempDir("dprint-plugin-gofumpt-cache-");
        try {
            const dprintConfig: Record<string, unknown> = {
                $schema: "https://dprint.dev/schemas/v0.json",
                gofumpt: {} as Record<string, unknown>,
                includes: ["test.go"],
                plugins: [path.resolve(WASM_FILE)],
            };
            const gofumptConfig = dprintConfig.gofumpt as Record<string, unknown>;
            if (tc.extraRules) {
                gofumptConfig.extraRules = true;
            }
            if (tc.extra.groupParams || tc.extra.clotheReturns) {
                gofumptConfig.extra = tc.extra;
            }
            if (tc.langVersion) {
                gofumptConfig.langVersion = tc.langVersion;
            }

            // If the archive has a go.mod, extract modulePath from it
            const goMod = fileMap.get("go.mod");
            if (goMod) {
                const moduleDirective = goMod.match(/^module\s+(\S+)/m);
                if (moduleDirective) {
                    gofumptConfig.modulePath = moduleDirective[1];
                }
            }

            await fs.promises.writeFile(
                path.join(testDir.dir, "dprint.json"),
                JSON.stringify(dprintConfig),
            );
            await fs.promises.writeFile(path.join(testDir.dir, "test.go"), inputData);

            const result = await $quiet({
                cwd: testDir.dir,
                env: { DPRINT_CACHE_DIR: cacheDir.dir },
                all: true,
            })`dprint fmt --log-level=debug --incremental=false`;

            try {
                const actual = await fs.promises.readFile(path.join(testDir.dir, "test.go"), "utf8");
                assert.strictEqual(
                    actual,
                    goldenData,
                    `Mismatch in ${path.basename(txtarPath)}: ${tc.inputFile} -> ${tc.goldenFile}`,
                );
            } catch (e) {
                process.stderr.write(result.all ?? "");
                throw e;
            }
        } finally {
            await testDir.cleanup();
            await cacheDir.cleanup();
        }
    }
}

export const test = task({
    name: "test",
    description: "Builds and runs tests.",
    dependencies: [build],
    run: async () => {
        const testDirs = await fs.promises.readdir("testdata");
        for (const testName of testDirs) {
            const stat = await fs.promises.stat(path.join("testdata", testName));
            if (!stat.isDirectory()) continue;
            console.log(`Running test: ${testName}`);
            await runTest(testName);
        }

        console.log(`\nRunning gofumpt txtar tests...`);
        const scriptDir = await getGofumptTestdataDir();
        const entries = await fs.promises.readdir(scriptDir);
        const txtarFiles = entries.filter((e) => e.endsWith(".txtar")).sort();

        let passed = 0;
        let skipped = 0;

        for (const file of txtarFiles) {
            const testName = file.replace(".txtar", "");
            if (skippedTxtarTests.has(testName)) {
                console.log(`Skipping: ${testName}`);
                skipped++;
                continue;
            }
            console.log(`Running: ${testName}`);
            await runGofumptTxtarTest(path.join(scriptDir, file));
            passed++;
        }

        console.log(`\nGofumpt tests: ${passed} passed, ${skipped} skipped`);
    },
});
