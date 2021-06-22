package walle_test

import (
    "fmt"
    "github.com/GGXXLL/walle"
)

func ExampleNewApk() {
    apk, err := walle.NewApk("test.apk")
    if err != nil {
        panic(err)
    }
    fmt.Println(apk.Path())
    fmt.Println(apk.Channel())
    fmt.Println(apk.Extras())
    newApk, err := apk.PutChannel("rrrr", "")
    if err != nil {
        panic(err)
    }
    fmt.Println(newApk.Path())
    fmt.Println(newApk.Channel())
    fmt.Println(newApk.Extras())
    // Output:
    // test.apk
    // rock
    // map[package_name:com.battery.cdyj version_code:10012]
    // test-rrrr.apk
    // rrrr
    // map[package_name:com.battery.cdyj version_code:10012]
}
