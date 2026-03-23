sed -i '1,2d' srcs/orchestration/service_test.go
awk '
/^import \($/ {
    print
    print "\t\"errors\""
    print "\t\"strings\""
    in_import = 1
    next
}
{print}
' srcs/orchestration/service_test.go > temp.go && mv temp.go srcs/orchestration/service_test.go
