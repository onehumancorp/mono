const fs = require('fs');
let content = fs.readFileSync('srcs/integrations/registry.go', 'utf8');

// Revert previous sed mistakes:
content = content.replace(/ValidateURL/g, "httputil.ValidateURL");
content = content.replace(/SafeClient/g, "httputil.SafeClient");

// Oh wait, if there are other places. Let's do it carefully.
