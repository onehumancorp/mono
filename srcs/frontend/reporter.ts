import type { Reporter, TestCase, TestResult, FullResult } from '@playwright/test/reporter';
import * as fs from 'fs';

class OHCGlassmorphismReporter implements Reporter {
  private results: { title: string; status: string; duration: number; error?: string }[] = [];

  onTestEnd(test: TestCase, result: TestResult) {
    this.results.push({
      title: test.title,
      status: result.status,
      duration: result.duration,
      error: result.error?.message,
    });
  }

  onEnd(result: FullResult) {
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>OHC Test Report</title>
    <style>
        body {
            font-family: 'Outfit', 'Inter', sans-serif;
            background: #0f172a;
            color: white;
            padding: 2rem;
            margin: 0;
            display: flex;
            justify-content: center;
        }
        .container {
            width: 80%;
            max-width: 1200px;
        }
        .header {
            text-align: center;
            margin-bottom: 2rem;
            font-size: 2rem;
            font-weight: bold;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1.5rem;
        }
        .card {
            backdrop-filter: blur(15px) saturate(180%);
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 1.5rem;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s ease-in-out;
        }
        .card:hover {
            transform: translateY(-5px);
        }
        .status-passed { color: #4ade80; }
        .status-failed { color: #f87171; }
        .status-timedOut { color: #facc15; }
        .title {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
        }
        .duration {
            font-size: 0.875rem;
            opacity: 0.7;
            margin-bottom: 1rem;
        }
        .error {
            background: rgba(248, 113, 113, 0.1);
            border-left: 3px solid #f87171;
            padding: 0.5rem;
            font-family: monospace;
            font-size: 0.875rem;
            overflow-x: auto;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">OHC Verification Report - ${result.status.toUpperCase()}</div>
        <div class="grid">
            ${this.results.map(r => `
                <div class="card">
                    <div class="title">${r.title}</div>
                    <div class="duration status-${r.status}">${r.status.toUpperCase()} (${r.duration}ms)</div>
                    ${r.error ? `<div class="error">${r.error.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</div>` : ''}
                </div>
            `).join('')}
        </div>
    </div>
</body>
</html>`;

    fs.writeFileSync('ohc-report.html', html);
    console.log('OHC Glassmorphism report generated at ohc-report.html');
  }
}

export default OHCGlassmorphismReporter;
