const { spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const protoc = path.resolve(process.argv[2]);
const plugin = path.resolve(process.argv[3]);
const proto = path.resolve(process.argv[4]);
const outDir = path.resolve(process.argv[5]);

console.log(`Resolved paths:`);
console.log(`  PROTOC: ${protoc}`);
console.log(`  PLUGIN: ${plugin}`);
console.log(`  PROTO:  ${proto}`);
console.log(`  OUTDIR: ${outDir}`);

function checkFile(p, name) {
  try {
    const stat = fs.statSync(p);
    console.log(`  ${name} exists, size: ${stat.size}, mode: ${stat.mode.toString(8)}`);
    return true;
  } catch (e) {
    console.error(`  ${name} NOT FOUND at ${p}: ${e.message}`);
    return false;
  }
}

const ok = checkFile(protoc, 'PROTOC') && checkFile(plugin, 'PLUGIN') && checkFile(proto, 'PROTO');

if (!ok) process.exit(1);

const args = [
  `--plugin=protoc-gen-ts=${plugin}`,
  `--ts_out=${outDir}`,
  `--ts_opt=esModuleInterop=true,forceLong=string,outputJsonMethods=false,outputClientImpl=false`,
  proto
];

console.log(`Executing: ${protoc} ${args.join(' ')}`);

const res = spawnSync(protoc, args, { stdio: 'inherit', env: process.env });

if (res.status !== 0) {
  console.error(`protoc failed with exit code ${res.status}, signal ${res.signal}`);
  if (res.error) console.error(`Error: ${res.error.message}`);
  process.exit(res.status || 1);
}

// Rename/move logic
function findHubTs(dir) {
  const files = fs.readdirSync(dir);
  for (const file of files) {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      const found = findHubTs(fullPath);
      if (found) return found;
    } else if (file === 'hub.ts') {
      return fullPath;
    }
  }
  return null;
}

const actualHubTs = findHubTs(outDir);
if (actualHubTs) {
  const targetHubTs = path.join(outDir, 'hub.ts');
  if (actualHubTs !== targetHubTs) {
    console.log(`Moving ${actualHubTs} -> ${targetHubTs}`);
    fs.renameSync(actualHubTs, targetHubTs);
  }
} else {
  console.error(`hub.ts not found in ${outDir}`);
  process.exit(1);
}

process.exit(0);
