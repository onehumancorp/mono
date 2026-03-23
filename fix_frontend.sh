sed -i 's/import { FormEvent, useEffect, useMemo, useState } from "react";/import { FormEvent, useEffect, useMemo, useState } from "react";\nimport { fetchMarketplace, importMarketplaceBlueprint, fetchAgents } from "\.\/api";/' srcs/frontend/src/App.tsx

sed -i 's/import type {/import type {\n  MarketplaceItem,\n  HandoffPackage,/' srcs/frontend/src/App.tsx
