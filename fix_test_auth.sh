sed -i 's/export VITE_BACKEND_URL/export ADMIN_USERNAME="admin"\nexport ADMIN_PASSWORD="adminpass123"\nexport VITE_BACKEND_URL/g' srcs/frontend/tests/unit_test.sh
sed -i 's/export VITE_BACKEND_URL/export ADMIN_USERNAME="admin"\nexport ADMIN_PASSWORD="adminpass123"\nexport VITE_BACKEND_URL/g' srcs/frontend/tests/vitest_test.sh
