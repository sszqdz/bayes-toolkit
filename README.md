# Bayes Toolkit

🚀 **Bayes Toolkit is a package written in Golang, aimed at facilitating rapid application development for users.**

> **NOTICE❗**  
> This toolkit is currently in the development phase, and more features will be added and enhanced in the future.

## 👉 Table of Contents

- [Installation](#-installation)
- [Packages](#-packages)
- [Examples](#-examples)
- [Repositories Directly Used in Bayes Toolkit](#-repositories-directly-used-in-bayes-toolkit)
- [License](#-license)

## 💻 Installation

First, use `go get`:

```shell
go get -u github.com/sszqdz/bayes-toolkit
```

Next, include in your code:

```go
import "github.com/sszqdz/bayes-toolkit"
```

## 📦 Packages

- **`ccmap`** Extend [![orcaman/concurrent-map](https://img.shields.io/github/stars/orcaman/concurrent-map?style=flat&color=blue&labelColor=black&label=orcaman/concurrent-map)](https://github.com/orcaman/concurrent-map) to conveniently allow **integer** type variables to be used as keys.
- **`dirr`** A complement to the standard library, providing **convenient file and directory operations** such as recursively searching for specific files up through parent directories.
- **`environment`** Load the **ENV** of the program (such as debug, release) from **system environment variables** (higher priority) or **.env** file.
- **`gin-router`** **A routing decorator** for [![gin-gonic/gin](https://img.shields.io/github/stars/gin-gonic/gin?style=flat&color=blue&labelColor=black&label=gin-gonic/gin)](https://github.com/gin-gonic/gin) that allows convenient **bypassing of specified middleware** for certain endpoints using methods such as **Unuse()** and **Skip()**.
- **`gin-timeout`** **A timeout middleware** for [![gin-gonic/gin](https://img.shields.io/github/stars/gin-gonic/gin?style=flat&color=blue&labelColor=black&label=gin-gonic/gin)](https://github.com/gin-gonic/gin) that **gracefully** handles timeout requests under **high concurrency**, **preventing Context leaks** and the resulting **request errors**.
- **`rrand`**  A complement to the standard library, providing **convenient operations for generating random values** such as random strings of a specified length.
- **`ws`** A wrapper for [![gorilla/websocket](https://img.shields.io/github/stars/gorilla/websocket?style=flat&color=blue&labelColor=black&label=gorilla/websocket)](https://github.com/gorilla/websocket), providing elegant and **safe concurrent read/write** operations, **graceful close and shutdown** handling.  

## 🗃 Examples

> 🚧 **Under Construction...**  

## 💖 Repositories Directly Used in Bayes Toolkit

**Hey, a big thumbs-up to all open source contributors out there!  
Your awesome work makes this boring world more interesting!  
Respect!**

- [![orcaman/concurrent-map](https://img.shields.io/github/stars/orcaman/concurrent-map?style=flat&color=blue&labelColor=black&label=orcaman/concurrent-map)](https://github.com/orcaman/concurrent-map) **a thread-safe concurrent map for go**
- [![spf13/cast](https://img.shields.io/github/stars/spf13/cast?style=flat&color=blue&labelColor=black&label=spf13/cast)](https://github.com/spf13/cast) **safe and easy casting from one type to another in Go**
- [![spf13/viper](https://img.shields.io/github/stars/spf13/viper?style=flat&color=blue&labelColor=black&label=spf13/viper)](https://github.com/spf13/viper) **Go configuration with fangs**
- [![gin-gonic/gin](https://img.shields.io/github/stars/gin-gonic/gin?style=flat&color=blue&labelColor=black&label=gin-gonic/gin)](https://github.com/gin-gonic/gin) **Gin is a HTTP web framework written in Go (Golang). It features a Martini-like API with much better performance -- up to 40 times faster. If you need smashing performance, get yourself some Gin.**
- [![gorilla/websocket](https://img.shields.io/github/stars/gorilla/websocket?style=flat&color=blue&labelColor=black&label=gorilla/websocket)](https://github.com/gorilla/websocket) **Package gorilla/websocket is a fast, well-tested and widely used WebSocket implementation for Go.**

AND these beautiful badges in SVG:  

- [![badges/shields](https://img.shields.io/github/stars/badges/shields?style=flat&color=blue&labelColor=black&label=badges/shields)](https://github.com/badges/shields) **Concise, consistent, and legible badges in SVG and raster format**  

## 📜 License

The project is licensed under the [MIT](https://github.com/sszqdz/bayes-toolkit/blob/master/LICENSE) © [sszqdz](https://github.com/sszqdz).
