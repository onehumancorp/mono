const fs = require('fs');

let content = fs.readFileSync('srcs/integration/frontend_backend_test.go', 'utf8');

// The file needs `import "github.com/onehumancorp/mono/srcs/httputil"`
// Let's find `import (` and insert it inside.

content = content.replace(/import \(/, 'import (\n\t"github.com/onehumancorp/mono/srcs/httputil"');
fs.writeFileSync('srcs/integration/frontend_backend_test.go', content);
