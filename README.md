# GoSnake
A Snake game that supports multiple players to play together online. The game is written in go language and does not use other third-party libraries except the standard library.

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
-  Grant permission 
```
chmod +x ./gosnake
```

- Run as a player
```
# play on my server
./gosnake

# or play on the server be deployed by yourself or others
./gosnake [-server-addr <game server address>]
```

When you don't specify the server-addr parameter, the server I deployed will be used. If you want to use your own server, then you need to run a server on the specified address like this：

- Run a game server
```
./gosnake -srv [-listen-addr <listen address>]
```

### Compile from source
Need to install go^1.15 and make tools
```
# build for current os
make build

# or build for all platforms (darwin, windows, linux)
make build_all
```
The builded file will be in the directory ./build/