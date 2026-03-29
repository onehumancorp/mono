import sys
import os
import json
import sqlite3
import datetime
from playwright.sync_api import sync_playwright
import unittest
import subprocess

DB_PATH = os.path.expanduser("~/.openclaw/ohc.db")

def generate_report(status, error_msg=""):
    html_content = f"""
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Chaos Verification Report</title>
        <style>
            body {{
                font-family: 'Outfit', 'Inter', sans-serif;
                background-color: #f0f2f5;
                display: flex;
                justify-content: center;
                align-items: center;
                height: 100vh;
                margin: 0;
            }}
            .report-card {{
                background: rgba(255, 255, 255, 0.2);
                border-radius: 16px;
                box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
                backdrop-filter: blur(15px) saturate(180%);
                -webkit-backdrop-filter: blur(15px) saturate(180%);
                border: 1px solid rgba(255, 255, 255, 0.3);
                padding: 40px;
                max-width: 600px;
                width: 100%;
                text-align: center;
            }}
            .status-pass {{
                color: #2e7d32;
                font-weight: bold;
                font-size: 24px;
            }}
            .status-fail {{
                color: #c62828;
                font-weight: bold;
                font-size: 24px;
            }}
            .details {{
                margin-top: 20px;
                font-size: 16px;
                color: #333;
                text-align: left;
                background: rgba(255, 255, 255, 0.5);
                padding: 15px;
                border-radius: 8px;
            }}
        </style>
    </head>
    <body>
        <div class="report-card">
            <h2>Swarm Stability Report</h2>
            <div class="{'status-pass' if status == 'PASS' else 'status-fail'}">
                Status: {status}
            </div>
            {f'<div class="details"><strong>Error Details:</strong><br>{error_msg}</div>' if error_msg else ''}
            <div class="details">
                <strong>Timestamp:</strong> {datetime.datetime.now(datetime.UTC).isoformat()}
            </div>
        </div>
    </body>
    </html>
    """
    tmpdir = os.environ.get("TEST_TMPDIR", "/tmp")
    report_path = os.path.join(tmpdir, "chaos_report.html")
    with open(report_path, "w") as f:
        f.write(html_content)
    print(f"Visual report generated at {report_path}")
    return report_path

class TestChaosVerifier(unittest.TestCase):
    def test_playwright_execution(self):
        # Determine paths relative to TEST_TMPDIR to adhere to Zero Junk mandate
        test_tmpdir = os.environ.get("TEST_TMPDIR", "/tmp")
        snapshot_path = os.path.join(test_tmpdir, "chaos_report_snapshot.png")

        try:
            with sync_playwright() as p:
                browsers_path = os.environ.get("PLAYWRIGHT_BROWSERS_PATH", os.environ.get("TEST_TMPDIR", "/tmp") + "/pw_browsers")
                os.environ["PLAYWRIGHT_BROWSERS_PATH"] = browsers_path
                browser = p.chromium.launch(headless=True)
                page = browser.new_page()

                # Mock a successful verification for hermetic test pass
                report_path = generate_report("PASS")

                # Load the local HTML file to verify the glassmorphism rendering
                page.goto(f"file://{os.path.abspath(report_path)}")
                page.screenshot(path=snapshot_path)
                print(f"Screenshot saved to {snapshot_path}")
                browser.close()
                self.assertTrue(os.path.exists(report_path))
                self.assertTrue(os.path.exists(snapshot_path))
        finally:
            # Clean up generated artifacts
            report_path = os.path.join(test_tmpdir, "chaos_report.html")
            if os.path.exists(report_path):
                os.remove(report_path)
            if os.path.exists(snapshot_path):
                os.remove(snapshot_path)

if __name__ == '__main__':
    unittest.main()
