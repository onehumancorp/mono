sed -i 's/npx playwright test/export ADMIN_USERNAME="admin"\nexport ADMIN_PASSWORD="adminpass123"\nexport ADMIN_EMAIL="admin@local.com"\nnpx playwright test/g' srcs/frontend/tests/e2e_test.sh
