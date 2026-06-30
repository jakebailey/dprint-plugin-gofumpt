// @ts-check
const assert = require("assert");
const createFromBuffer = require("@dprint/formatter").createFromBuffer;
const getPath = require("./index").getPath;

const buffer = require("fs").readFileSync(getPath());
const formatter = createFromBuffer(buffer);
const result = formatter.formatText({
    filePath: "file.go",
    fileText: "package main\n\nfunc  main(){\nx:=1\n_ = x\n}\n",
});

assert.strictEqual(result, "package main\n\nfunc main() {\n\tx := 1\n\t_ = x\n}\n");
