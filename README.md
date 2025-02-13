## **🚀 Full Inventory Management System: Golang + SQLite + JWT + Docker + Secret Key Management**
This guide will give you **everything**—from setting up your environment, writing the backend API, creating a frontend, implementing JWT authentication with **secure key storage**, and **Dockerizing the application**.

---

## **📌 1️⃣ Setup Your Development Environment**
### **Install Go**
1. Download and install **Go** → [https://go.dev/dl/](https://go.dev/dl/)
2. Verify installation:
   ```bash
   go version
   ```

### **Set Up Your GitHub Repo**
1. Create a **GitHub repository** (e.g., `inventory-management-go`).
2. Clone it into VS Code:
   ```bash
   git clone https://github.com/your-username/inventory-management-go.git
   cd inventory-management-go
   ```
3. Initialize **Go Modules**:
   ```bash
   go mod init github.com/your-username/inventory-management-go
   ```

---

## **📌 2️⃣ Install Dependencies**
Run:
```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/gin-contrib/cors
go get -u github.com/gin-contrib/limiter
go get -u github.com/joho/godotenv
```

---

## **📌 3️⃣ Securely Store and Generate Secret Key**
Instead of hardcoding the secret key, we store it securely in **`.env`**.

### **🔹 Create a `.env` File**
```bash
touch .env
echo "JWT_SECRET_KEY=$(openssl rand -hex 32)" >> .env
```

### **🔹 Load Secret Key in Go (`config.go`)**
```go
package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

var SecretKey string

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    SecretKey = os.Getenv("JWT_SECRET_KEY")
    if SecretKey == "" {
        log.Fatal("JWT_SECRET_KEY is missing in .env")
    }
}
```

---

## **📌 4️⃣ Database Setup (`database.go`)**
```go
package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
)

var DB *sql.DB

func InitDB() {
    var err error
    DB, err = sql.Open("sqlite3", "inventory.db")
    if err != nil {
        log.Fatal(err)
    }

    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS inventory (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            price FLOAT NOT NULL
        );
    `)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## **📌 5️⃣ JWT Authentication (`auth.go`)**
```go
package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "net/http"
    "time"
    "inventory_app/config"
)

func GenerateToken(username string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "exp":      time.Now().Add(time.Hour * 72).Unix(),
    })
    return token.SignedString([]byte(config.SecretKey))
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(config.SecretKey), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## **📌 6️⃣ API with Pagination & Fuzzy Search (`main.go`)**
```go
package main

import (
    "github.com/gin-gonic/gin"
    "inventory_app/database"
    "inventory_app/auth"
    "inventory_app/config"
    "strconv"
)

func main() {
    config.LoadConfig()
    database.InitDB()
    
    r := gin.Default()
    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")

    r.GET("/", Home)
    r.POST("/signup", Signup)
    r.POST("/login", Login)

    protected := r.Group("/")
    protected.Use(auth.AuthMiddleware())
    protected.GET("/items", GetItems)
    protected.POST("/items", CreateItem)
    protected.DELETE("/items/:id", DeleteItem)

    r.Run(":8080")
}

func Home(c *gin.Context) {
    c.HTML(200, "index.html", nil)
}

func GetItems(c *gin.Context) {
    searchQuery := c.Query("q")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit := 5
    offset := (page - 1) * limit

    var rows *sql.Rows
    var err error

    if searchQuery != "" {
        rows, err = database.DB.Query("SELECT id, name, description, price FROM inventory WHERE name LIKE ? OR description LIKE ? LIMIT ? OFFSET ?", "%"+searchQuery+"%", "%"+searchQuery+"%", limit, offset)
    } else {
        rows, err = database.DB.Query("SELECT id, name, description, price FROM inventory LIMIT ? OFFSET ?", limit, offset)
    }

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var items []map[string]interface{}
    for rows.Next() {
        var id int
        var name, description string
        var price float64
        rows.Scan(&id, &name, &description, &price)
        items = append(items, gin.H{"id": id, "name": name, "description": description, "price": price})
    }

    c.JSON(200, gin.H{"items": items, "page": page})
}
```

---

## **📌 7️⃣ Frontend (`templates/index.html`)**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Inventory Management</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css">
    <script>
        function loadItems(page = 1) {
            let query = document.getElementById("search").value;
            let url = `/items?page=${page}`;
            if (query) url += `&q=${query}`;

            fetch(url)
                .then(response => response.json())
                .then(data => {
                    let tableBody = document.getElementById("inventoryTable");
                    tableBody.innerHTML = "";
                    data.items.forEach(item => {
                        tableBody.innerHTML += `<tr><td>${item.id}</td><td>${item.name}</td><td>${item.description}</td><td>${item.price}</td></tr>`;
                    });
                });
        }

        window.onload = function() {
            loadItems();
        };
    </script>
</head>
<body>
    <div class="container">
        <h1>Inventory Management</h1>
        <input type="text" id="search" placeholder="Search Items" onkeyup="loadItems(1)">
        <table class="table"><thead><tr><th>ID</th><th>Name</th><th>Description</th><th>Price</th></tr></thead><tbody id="inventoryTable"></tbody></table>
    </div>
</body>
</html>
```

---

## **📌 8️⃣ Docker Setup**
### **🔹 `Dockerfile`**
```dockerfile
FROM golang:1.20
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o inventory_app
CMD ["./inventory_app"]
EXPOSE 8080
```

### **🔹 Run with Docker**
```bash
docker build -t inventory-app .
docker run -p 8080:8080 --env-file .env inventory-app
```

---

## **🚀 Your Complete Setup is Ready!**
Would you like **AWS deployment** next? 🚀🔥## **🚀 Full Inventory Management System: Golang + SQLite + JWT + Docker + Secret Key Management**
This guide will give you **everything**—from setting up your environment, writing the backend API, creating a frontend, implementing JWT authentication with **secure key storage**, and **Dockerizing the application**.

---

## **📌 1️⃣ Setup Your Development Environment**
### **Install Go**
1. Download and install **Go** → [https://go.dev/dl/](https://go.dev/dl/)
2. Verify installation:
   ```bash
   go version
   ```

### **Set Up Your GitHub Repo**
1. Create a **GitHub repository** (e.g., `inventory-management-go`).
2. Clone it into VS Code:
   ```bash
   git clone https://github.com/your-username/inventory-management-go.git
   cd inventory-management-go
   ```
3. Initialize **Go Modules**:
   ```bash
   go mod init github.com/your-username/inventory-management-go
   ```

---

## **📌 2️⃣ Install Dependencies**
Run:
```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/gin-contrib/cors
go get -u github.com/gin-contrib/limiter
go get -u github.com/joho/godotenv
```

---

## **📌 3️⃣ Securely Store and Generate Secret Key**
Instead of hardcoding the secret key, we store it securely in **`.env`**.

### **🔹 Create a `.env` File**
```bash
touch .env
echo "JWT_SECRET_KEY=$(openssl rand -hex 32)" >> .env
```

### **🔹 Load Secret Key in Go (`config.go`)**
```go
package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

var SecretKey string

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    SecretKey = os.Getenv("JWT_SECRET_KEY")
    if SecretKey == "" {
        log.Fatal("JWT_SECRET_KEY is missing in .env")
    }
}
```

---

## **📌 4️⃣ Database Setup (`database.go`)**
```go
package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
)

var DB *sql.DB

func InitDB() {
    var err error
    DB, err = sql.Open("sqlite3", "inventory.db")
    if err != nil {
        log.Fatal(err)
    }

    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS inventory (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            price FLOAT NOT NULL
        );
    `)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## **📌 5️⃣ JWT Authentication (`auth.go`)**
```go
package auth

import (
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "net/http"
    "time"
    "inventory_app/config"
)

func GenerateToken(username string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "exp":      time.Now().Add(time.Hour * 72).Unix(),
    })
    return token.SignedString([]byte(config.SecretKey))
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(config.SecretKey), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## **📌 6️⃣ API with Pagination & Fuzzy Search (`main.go`)**
```go
package main

import (
    "github.com/gin-gonic/gin"
    "inventory_app/database"
    "inventory_app/auth"
    "inventory_app/config"
    "strconv"
)

func main() {
    config.LoadConfig()
    database.InitDB()
    
    r := gin.Default()
    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")

    r.GET("/", Home)
    r.POST("/signup", Signup)
    r.POST("/login", Login)

    protected := r.Group("/")
    protected.Use(auth.AuthMiddleware())
    protected.GET("/items", GetItems)
    protected.POST("/items", CreateItem)
    protected.DELETE("/items/:id", DeleteItem)

    r.Run(":8080")
}

func Home(c *gin.Context) {
    c.HTML(200, "index.html", nil)
}

func GetItems(c *gin.Context) {
    searchQuery := c.Query("q")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit := 5
    offset := (page - 1) * limit

    var rows *sql.Rows
    var err error

    if searchQuery != "" {
        rows, err = database.DB.Query("SELECT id, name, description, price FROM inventory WHERE name LIKE ? OR description LIKE ? LIMIT ? OFFSET ?", "%"+searchQuery+"%", "%"+searchQuery+"%", limit, offset)
    } else {
        rows, err = database.DB.Query("SELECT id, name, description, price FROM inventory LIMIT ? OFFSET ?", limit, offset)
    }

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var items []map[string]interface{}
    for rows.Next() {
        var id int
        var name, description string
        var price float64
        rows.Scan(&id, &name, &description, &price)
        items = append(items, gin.H{"id": id, "name": name, "description": description, "price": price})
    }

    c.JSON(200, gin.H{"items": items, "page": page})
}
```

---

## **📌 7️⃣ Frontend (`templates/index.html`)**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Inventory Management</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css">
    <script>
        function loadItems(page = 1) {
            let query = document.getElementById("search").value;
            let url = `/items?page=${page}`;
            if (query) url += `&q=${query}`;

            fetch(url)
                .then(response => response.json())
                .then(data => {
                    let tableBody = document.getElementById("inventoryTable");
                    tableBody.innerHTML = "";
                    data.items.forEach(item => {
                        tableBody.innerHTML += `<tr><td>${item.id}</td><td>${item.name}</td><td>${item.description}</td><td>${item.price}</td></tr>`;
                    });
                });
        }

        window.onload = function() {
            loadItems();
        };
    </script>
</head>
<body>
    <div class="container">
        <h1>Inventory Management</h1>
        <input type="text" id="search" placeholder="Search Items" onkeyup="loadItems(1)">
        <table class="table"><thead><tr><th>ID</th><th>Name</th><th>Description</th><th>Price</th></tr></thead><tbody id="inventoryTable"></tbody></table>
    </div>
</body>
</html>
```

---

## **📌 8️⃣ Docker Setup**
### **🔹 `Dockerfile`**
```dockerfile
FROM golang:1.20
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o inventory_app
CMD ["./inventory_app"]
EXPOSE 8080
```

### **🔹 Run with Docker**
```bash
docker build -t inventory-app .
docker run -p 8080:8080 --env-file .env inventory-app
```

---

## **🚀 Your Complete Setup is Ready!**
