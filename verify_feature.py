from playwright.sync_api import sync_playwright

def verify_feature(page):
    page.goto("http://localhost:5173")
    page.wait_for_timeout(2000)

    try:
        page.locator("input[id='login-username']").fill("admin")
        page.locator("input[id='login-password']").fill("admin")
        page.get_by_role("button", name="Sign in").click()
    except Exception as e:
        print(f"Login failed/skipped: {e}")

    page.wait_for_timeout(2000)

    # Click navigation item for scaling
    page.get_by_role("button", name="Dynamic Scaling").click()
    page.wait_for_timeout(1000)

    # Wait for the sliders to render
    page.wait_for_selector(".scaling-slider")

    # Take screenshot of the scaling page
    page.screenshot(path="/home/jules/verification/verification.png")
    page.wait_for_timeout(1000)

    # Adjust the scaling
    slider = page.locator("input[type=range]").nth(1) # SWE
    slider.focus()
    for _ in range(2):
        page.keyboard.press("ArrowRight")

    page.wait_for_timeout(500)

    page.get_by_role("button", name="Apply Scaling Changes").click()
    page.wait_for_timeout(3000)

    page.screenshot(path="/home/jules/verification/after_scaling.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    import os
    os.makedirs("/home/jules/verification/video", exist_ok=True)
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(record_video_dir="/home/jules/verification/video")
        page = context.new_page()
        try:
            verify_feature(page)
        finally:
            context.close()
            browser.close()
