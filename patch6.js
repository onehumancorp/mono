const fs = require('fs');

let content = fs.readFileSync('srcs/integrations/registry.go', 'utf8');

// I need to add httputil import to registry.go since it uses it in other places? Let's check where else it uses it.
