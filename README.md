Readme
=======
This is a console douban client for windows, the most difference from others is 
that it supports `global hotkey`, which is cool for people who enjoy keyboard 
(like me, a vim fans).  
By using `global hotkey`, means you can manipulate the client without having 
to make the client window focus, I think it is definitely convenient when you 
are working.  

Anyway, it's a toy written by `Golang` during I learn `Golang`. :-)

### How to use it
+ Download mpg123 player and unzip it to some directory  
+ Change config info in `config.cfg`  
    - `USER` means your accout in douban  
    - `PASSWD` means your password in douban  
    - `MPG123_DIR` means the directory of `mpg123` you have downloaded and
    unzip, `mpg123.exe` must be in the top directory of `MPG123_DIR`
    
+ See usage by typing `ctrl+shift+F9` after click `hotkeysdouban.exe`

### How to build  
+ Build with an icon  

    rsrc -manifest hotkeysdouban.exe.manifest -ico hotkeysdouban.ico -o rsrc.syso  
    go build

+ Or Build it only without an icon

    go build

### Thanks
Have a nice day. :-)
