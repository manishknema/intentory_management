package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/manishknema/inventory_management/database"
	"github.com/manishknema/inventory_management/models"
)

// GetItems retrieves all inventory items with pagination
func GetItems(c *gin.Context) {
	log.Println("📥 Received request to GetItems")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		log.Println("❌ Invalid page number:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	pageSize := 5
	offset := (page - 1) * pageSize

	log.Printf("🔍 Fetching items from database (LIMIT %d OFFSET %d)", pageSize, offset)
	rows, err := database.DB.Query("SELECT id, name, description, price FROM inventory LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		log.Println("❌ SQL Query Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price)
		if err != nil {
			log.Println("❌ Error scanning row:", err)
			continue
		}
		items = append(items, item)
	}

	log.Printf("✅ Retrieved %d items", len(items))
	c.JSON(http.StatusOK, gin.H{"items": items, "page": page})
}

// GetItem retrieves a single item by ID
func GetItem(c *gin.Context) {
	id := c.Param("id")
	log.Println("📥 Received request to GetItem with ID:", id)

	var item models.Item
	err := database.DB.QueryRow("SELECT id, name, description, price FROM inventory WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Price)

	if err == sql.ErrNoRows {
		log.Println("❌ Item not found with ID:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	} else if err != nil {
		log.Println("❌ SQL Query Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("✅ Retrieved item:", item)
	c.JSON(http.StatusOK, item)
}

// CreateItem adds a new item to the inventory
func CreateItem(c *gin.Context) {
	log.Println("📥 Received request to CreateItem")

	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		log.Println("❌ Error parsing request JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate that price is a float
	if item.Price <= 0 {
		log.Println("❌ Invalid price value:", item.Price)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be a positive number"})
		return
	}

	log.Println("🔍 Inserting item:", item)
	result, err := database.DB.Exec("INSERT INTO inventory (name, description, price) VALUES (?, ?, ?)",
		item.Name, item.Description, item.Price)
	if err != nil {
		log.Println("❌ SQL Insert Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create item"})
		return
	}

	lastInsertID, _ := result.LastInsertId()
	log.Println("✅ Item Inserted with ID:", lastInsertID)
	c.JSON(http.StatusCreated, gin.H{"message": "Item created successfully", "id": lastInsertID})
}

// UpdateItem modifies an existing item
func UpdateItem(c *gin.Context) {
	id := c.Param("id")
	log.Println("📥 Received request to UpdateItem with ID:", id)

	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		log.Println("❌ Error parsing request JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate that price is a float
	if item.Price <= 0 {
		log.Println("❌ Invalid price value:", item.Price)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be a positive number"})
		return
	}

	log.Println("🔍 Updating item:", item)
	result, err := database.DB.Exec("UPDATE inventory SET name = ?, description = ?, price = ? WHERE id = ?",
		item.Name, item.Description, item.Price, id)
	if err != nil {
		log.Println("❌ SQL Update Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update item"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Println("⚠️ No rows updated for ID:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	log.Println("✅ Item updated successfully:", id)
	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully"})
}

// DeleteItems removes multiple items from the inventory
func DeleteItems(c *gin.Context) {
	log.Println("📥 Received request to DeleteItems")

	var request struct {
		ItemIDs []int `json:"item_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("❌ Error parsing request JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if len(request.ItemIDs) == 0 {
		log.Println("⚠️ No items selected for deletion")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No items selected for deletion"})
		return
	}

	placeholders := strings.Repeat("?,", len(request.ItemIDs)-1) + "?"
	query := "DELETE FROM inventory WHERE id IN (" + placeholders + ")"

	args := make([]interface{}, len(request.ItemIDs))
	for i, id := range request.ItemIDs {
		args[i] = id
	}

	log.Println("🔍 Deleting items:", request.ItemIDs)
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		log.Println("❌ SQL Delete Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete items"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Println("⚠️ No items deleted")
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching items found for deletion"})
		return
	}

	log.Printf("✅ %d items deleted successfully", rowsAffected)
	c.JSON(http.StatusOK, gin.H{"message": "Selected items deleted successfully", "deleted_count": rowsAffected})
}
