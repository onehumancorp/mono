import "@testing-library/jest-dom/vitest";

const backendUrl = process.env.VITE_BACKEND_URL || "http://127.0.0.1:8080";
const originalFetch = globalThis.fetch;

globalThis.fetch = async (input, init) => {
  if (typeof input === "string" && input.startsWith("/api/")) {
    let tries = 5;
    while (tries > 0) {
        try {
            return await originalFetch(`${backendUrl}${input}`, init);
        } catch (e: any) {
            if (e.code === 'ECONNREFUSED' && tries > 1) {
                tries--;
                await new Promise(r => setTimeout(r, 1000));
            } else {
                throw e;
            }
        }
    }
  }
  return originalFetch(input, init);
};
