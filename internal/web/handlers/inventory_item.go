package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/domain"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type InventoryItemManager struct {
	InventoryItemRepo data.InventoryItemRepo
}

// RegisterInventoryItemRoutes wires inventory item management endpoints under /app.
func RegisterInventoryItemRoutes(group *ghttp.RouterGroup, inventoryItemRepo data.InventoryItemRepo) {
	iim := &InventoryItemManager{
		InventoryItemRepo: inventoryItemRepo,
	}

	// InventoryItem management
	group.GET("/management/inventory-items", iim.InventoryItemsGet)
	group.POST("/management/inventory-items", iim.InventoryItemPost)
	group.GET("/management/inventory-items/new", iim.InventoryItemGet)
	group.GET("/management/inventory-items/:id", iim.InventoryItemGet)
	group.PUT("/management/inventory-items/:id", iim.InventoryItemPut)
	group.DELETE("/management/inventory-items/:id", iim.InventoryItemDelete)
}

// InventoryItemsGet renders the inventory items management page.
func (iim *InventoryItemManager) InventoryItemsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	inventoryItems, err := iim.InventoryItemRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list inventory items: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.InventoryItemsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				inventoryItems,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.InventoryItemsPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			inventoryItems,
		),
	)
}

// InventoryItemPost creates a new inventory item.
func (iim *InventoryItemManager) InventoryItemPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	itemType := strings.TrimSpace(r.Get("type").String())
	quantityStr := strings.TrimSpace(r.Get("quantity").String())
	unit := strings.TrimSpace(r.Get("unit").String())
	expirationDateStr := strings.TrimSpace(r.Get("expiration_date").String())
	supplierInfo := strings.TrimSpace(r.Get("supplier_info").String())
	notes := strings.TrimSpace(r.Get("notes").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var typePtr *string
	if itemType != "" {
		typePtr = new(string)
		*typePtr = itemType
	}

	var quantity *float64
	if quantityStr != "" {
		if qty, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = new(float64)
			*quantity = qty
		} else {
			errs["quantity"] = "Quantity must be a valid number"
		}
	}

	var unitPtr *string
	if unit != "" {
		unitPtr = new(string)
		*unitPtr = unit
	}

	var expirationDate *time.Time
	if expirationDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", expirationDateStr); err == nil {
			expirationDate = new(time.Time)
			*expirationDate = parsedDate
		} else {
			errs["expiration_date"] = "Expiration date must be in YYYY-MM-DD format"
		}
	}

	var supplierPtr *string
	if supplierInfo != "" {
		supplierPtr = new(string)
		*supplierPtr = supplierInfo
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = new(string)
		*notesPtr = notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		inventoryItem := &domain.InventoryItem{
			Name:           name,
			Type:           typePtr,
			Quantity:       quantity,
			Unit:           unitPtr,
			ExpirationDate: expirationDate,
			SupplierInfo:   supplierPtr,
			Notes:          notesPtr,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := iim.InventoryItemRepo.Create(r.GetCtx(), inventoryItem)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create inventory item: %v", err)
			errs["form"] = "Failed to create inventory item"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(middleware.BasePath() + "/management/inventory-items")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/inventory-items")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/inventory-items")
}

// InventoryItemGet renders a specific inventory item for editing or a new inventory item form.
func (iim *InventoryItemManager) InventoryItemGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var inventoryItem *domain.InventoryItem

	// Check if this is a request for a new inventory item (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new inventory item
		inventoryItem = nil
	} else {
		// This is a request for editing an existing inventory item
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid inventory item ID")
			return
		}

		inventoryItem, err = iim.InventoryItemRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Inventory item not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find inventory item: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.InventoryItemContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				inventoryItem,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.InventoryItemPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			ThemeToString(user.Theme),
			inventoryItem,
		),
	)
}

// InventoryItemPut updates an existing inventory item.
func (iim *InventoryItemManager) InventoryItemPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid inventory item ID")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	itemType := strings.TrimSpace(r.Get("type").String())
	quantityStr := strings.TrimSpace(r.Get("quantity").String())
	unit := strings.TrimSpace(r.Get("unit").String())
	expirationDateStr := strings.TrimSpace(r.Get("expiration_date").String())
	supplierInfo := strings.TrimSpace(r.Get("supplier_info").String())
	notes := strings.TrimSpace(r.Get("notes").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var typePtr *string
	if itemType != "" {
		typePtr = new(string)
		*typePtr = itemType
	}

	var quantity *float64
	if quantityStr != "" {
		if qty, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = new(float64)
			*quantity = qty
		} else {
			errs["quantity"] = "Quantity must be a valid number"
		}
	}

	var unitPtr *string
	if unit != "" {
		unitPtr = new(string)
		*unitPtr = unit
	}

	var expirationDate *time.Time
	if expirationDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", expirationDateStr); err == nil {
			expirationDate = new(time.Time)
			*expirationDate = parsedDate
		} else {
			errs["expiration_date"] = "Expiration date must be in YYYY-MM-DD format"
		}
	}

	var supplierPtr *string
	if supplierInfo != "" {
		supplierPtr = new(string)
		*supplierPtr = supplierInfo
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = new(string)
		*notesPtr = notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		inventoryItem := &domain.InventoryItem{
			InventoryItemID: id,
			Name:            name,
			Type:            typePtr,
			Quantity:        quantity,
			Unit:            unitPtr,
			ExpirationDate:  expirationDate,
			SupplierInfo:    supplierPtr,
			Notes:           notesPtr,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := iim.InventoryItemRepo.Update(r.GetCtx(), inventoryItem)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update inventory item: %v", err)
			errs["form"] = "Failed to update inventory item"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(fmt.Sprintf("%s/management/inventory-items/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/inventory-items/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated inventory item
	r.Response.RedirectTo(fmt.Sprintf("%s/management/inventory-items/%d", middleware.BasePath(), id))
}

// InventoryItemDelete soft deletes an inventory item.
func (iim *InventoryItemManager) InventoryItemDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid inventory item ID")
		return
	}

	err = iim.InventoryItemRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete inventory item: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/inventory-items")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/inventory-items")
}
