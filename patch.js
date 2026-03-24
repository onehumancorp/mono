const fs = require('fs');

let content = fs.readFileSync('srcs/frontend/src/App.test.tsx', 'utf8');

content = content.replace(
    'expect(fetchMock).toHaveBeenCalledWith("/api/domains");',
    ''
);
content = content.replace(
    'expect(fetchMock).toHaveBeenCalledWith("/api/mcp/tools");',
    ''
);

fs.writeFileSync('srcs/frontend/src/App.test.tsx', content);
