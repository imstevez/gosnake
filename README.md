# gosnake
A classical snake game written in go language. No 3rd party library dependencies.

<image src="https://raw.githubusercontent.com/imstevez/gosnake/main/game_show.gif" width="300px">

### Run the binary directly
- Download the corresponding version of the executable file
```
# on macos
curl https://raw.githubusercontent.com/imstevez/gosnacks/main/build/gosnake_darwin > ./gosnake
# on windows
curl https://raw.githubusercontent.com/imstevez/gosnacks/main/build/gosnake_windows > ./gosnake
# on linux
curl https://raw.githubusercontent.com/imstevez/gosnacks/main/build/gosnake_linux > ./gosnake
```
-  Grant permission and execute
```
chmod +x ./gosnake
./gosnake
```

### Compile from source
Need to install go^1.15 and make tools
```
make build
```
The builded file will be ./build/gosnake