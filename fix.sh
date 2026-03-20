sed -i 's/app, _ := newTestServer(t)/app, _, _ := newTestServer(t)/g' srcs/dashboard/server_missing_test.go
sed -i 's/_, server := newTestServer(t)/_, server, _ := newTestServer(t)/g' srcs/dashboard/server_missing_test.go
sed -i 's/app, server := newTestServer(t)/app, server, _ := newTestServer(t)/g' srcs/dashboard/server_missing_test.go
