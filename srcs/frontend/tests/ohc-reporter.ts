import type { Reporter, FullConfig, Suite, TestCase, TestResult, FullResult } from '@playwright/test/reporter';
import * as fs from 'fs';
import * as path from 'path';

class OHCReporter implements Reporter {
  private outputDir: string = 'playwright-report';
  private testResults: { title: string; status: string; duration: number; error?: string }[] = [];

  onBegin(config: FullConfig, suite: Suite) {
    console.log(`Starting the run with ${suite.allTests().length} tests`);
  }

  onTestBegin(test: TestCase, result: TestResult) {
    console.log(`Starting test ${test.title}`);
  }

  onTestEnd(test: TestCase, result: TestResult) {
    this.testResults.push({
      title: test.title,
      status: result.status,
      duration: result.duration,
      error: result.error?.message,
    });
    console.log(`Finished test ${test.title}: ${result.status}`);
  }

  async onEnd(result: FullResult) {
    console.log(`Finished the run: ${result.status}`);
    this.generateHtmlReport(result);
  }

  private generateHtmlReport(result: FullResult) {
    const reportPath = path.join(this.outputDir, 'ohc-report.html');
    if (!fs.existsSync(this.outputDir)) {
      fs.mkdirSync(this.outputDir, { recursive: true });
    }

    let cardsHtml = '';
    for (const test of this.testResults) {
        let errorHtml = '';
        if (test.error) {
            const escapedError = test.error.replace(/</g, '&lt;').replace(/>/g, '&gt;');
            errorHtml = `<div class="error-message"><pre>${escapedError}</pre></div>`;
        }

        cardsHtml += `
            <div class="card ${test.status}">
                <h3 class="test-title">${test.title}</h3>
                <div class="test-meta">
                    Duration: ${(test.duration / 1000).toFixed(2)}s
                </div>
                <span class="status-badge ${test.status}">${test.status}</span>
                ${errorHtml}
            </div>
        `;
    }

    const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OHC Test Report</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Outfit:wght@400;500;600;700&display=swap');

        body {
            font-family: 'Outfit', 'Inter', sans-serif;
            background-color: #0f172a;
            color: #f8fafc;
            margin: 0;
            padding: 2rem;
            min-height: 100vh;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        .header {
            margin-bottom: 2rem;
        }

        .header h1 {
            font-size: 2.5rem;
            font-weight: 700;
            margin: 0;
            background: linear-gradient(to right, #60a5fa, #a78bfa);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1.5rem;
        }

        .card {
            backdrop-filter: blur(15px) saturate(180%);
            -webkit-backdrop-filter: blur(15px) saturate(180%);
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 1rem;
            padding: 1.5rem;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }

        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5);
        }

        .card.passed {
            border-left: 4px solid #34d399;
        }

        .card.failed {
            border-left: 4px solid #ef4444;
        }

        .card.timedOut {
            border-left: 4px solid #f59e0b;
        }

        .test-title {
            font-size: 1.125rem;
            font-weight: 600;
            margin: 0 0 0.5rem 0;
        }

        .test-meta {
            font-family: 'Inter', sans-serif;
            font-size: 0.875rem;
            color: #94a3b8;
            margin-bottom: 1rem;
        }

        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .status-badge.passed {
            background-color: rgba(52, 211, 153, 0.1);
            color: #34d399;
        }

        .status-badge.failed {
            background-color: rgba(239, 68, 68, 0.1);
            color: #ef4444;
        }

        .error-message {
            font-family: 'Inter', monospace;
            font-size: 0.875rem;
            background: rgba(0, 0, 0, 0.3);
            padding: 1rem;
            border-radius: 0.5rem;
            color: #ef4444;
            overflow-x: auto;
            margin-top: 1rem;
        }

        .summary {
            display: flex;
            gap: 2rem;
            margin-bottom: 2rem;
            padding: 1.5rem;
            backdrop-filter: blur(15px) saturate(180%);
            -webkit-backdrop-filter: blur(15px) saturate(180%);
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 1rem;
        }

        .stat-group {
            display: flex;
            flex-direction: column;
        }

        .stat-label {
            font-family: 'Inter', sans-serif;
            font-size: 0.875rem;
            color: #94a3b8;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .stat-value {
            font-size: 2rem;
            font-weight: 700;
        }

        .stat-value.passed { color: #34d399; }
        .stat-value.failed { color: #ef4444; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Swarm Verification Report</h1>
        </div>

        <div class="summary">
            <div class="stat-group">
                <span class="stat-label">Total Tests</span>
                <span class="stat-value">${this.testResults.length}</span>
            </div>
            <div class="stat-group">
                <span class="stat-label">Passed</span>
                <span class="stat-value passed">${this.testResults.filter(t => t.status === 'passed').length}</span>
            </div>
            <div class="stat-group">
                <span class="stat-label">Failed</span>
                <span class="stat-value failed">${this.testResults.filter(t => t.status !== 'passed' && t.status !== 'skipped').length}</span>
            </div>
        </div>

        <div class="grid">
            ${cardsHtml}
        </div>
    </div>
</body>
</html>
    `;

    fs.writeFileSync(reportPath, htmlContent);
    console.log(`Report generated at: ${reportPath}`);
  }
}

export default OHCReporter;
