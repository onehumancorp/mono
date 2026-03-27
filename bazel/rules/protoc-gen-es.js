const { runNodeJs } = require("@bufbuild/protoplugin");
const { protocGenEs } = require("@bufbuild/protoc-gen-es");

runNodeJs(protocGenEs);
