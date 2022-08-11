# go-mobile

go-mobile is a set of tools for using Go on mobile platforms, in order to embeed it on smartphone applications.

Ultrablue uses it because some libraries it needs aren't available on IOS/Android's native plateform.

## Code restrictions

The following points are rather a collection of advice I wish I had known when I first used gomobile than strong requirements.

- Functions must return pointers
- You can return structures
- You can't return nested structures
- You can't return interface types / neither structures with interface fields
- Byte arrays are'nt handled very well, thus prefer a structure with a unique byte array field
- error is the only type you can return without a pointer to it. If an error is returned, it will raise an exception that the caller will be able to handle, not returning him an error value.
- To make a structure field available to the caller, you must export it as you would do for any go package.

For more reliable information, please refer to the gomobile documentation.


## Usage

Once you've written your go code, here's how to use it:

**Install gomobile:**
```
go install golang.org/x/mobile/cmd/gobind@latest
go install golang.org/x/mobile/cmd/gomobile@latest
go mod download golang.org/x/mobile
go get golang.org/x/mobile/bind
```
Make sure it is on your `PATH` once installed.

**Set needed environment variables:**
```
export ANDROID_HOME=/path/to/Sdk
export ANDROID_NDK_HOME=/path/to/ndk-bundle
```

**Compile your code to the desired architecture:**
```
gomobile bind -target=ios -v . #Note that Xcode is required
```
or
```
gomobile bind -target=android -v .
```
This will produce an `.aar` archive for Android, or a `.XCFramework` for IOS, please refer to specific documentation to include those in your project.
