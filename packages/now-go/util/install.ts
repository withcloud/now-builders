import { join } from 'path';
import { downloadGo } from '../go-helpers';

async function main() {
  // First download the `go` binary for this platform/arch.
  const go = await downloadGo();

  // Build the `analyse` helper program.
  // `go get` is not necessary because the program has no external deps.
  const src = join(__dirname, 'analyse.go');
  const dest = join(__dirname, '../analyse');
  await go.build(src, dest);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
