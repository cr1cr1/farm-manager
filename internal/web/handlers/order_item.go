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

type OrderItemManager struct {
	OrderItemRepo data.OrderItemRepo
	OrderRepo     data.OrderRepo
}

// RegisterOrderItemRoutes wires order item management endpoints under /app.
func RegisterOrderItemRoutes(group *ghttp.RouterGroup, orderItemRepo data.OrderItemRepo, orderRepo data.OrderRepo) {
	oim := &OrderItemManager{
		OrderItemRepo: orderItemRepo,
		OrderRepo:     orderRepo,
	}

	// Order item management
	group.GET("/management/order-items", oim.OrderItemsGet)
	group.POST("/management/order-items", oim.OrderItemPost)
	group.GET("/management/order-items/new", oim.OrderItemGet)
	group.GET("/management/order-items/:id", oim.OrderItemGet)
	group.PUT("/management/order-items/:id", oim.OrderItemPut)
	group.DELETE("/management/order-items/:id", oim.OrderItemDelete)
}

// OrderItemsGet renders the order items management page.
func (oim *OrderItemManager) OrderItemsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	orderItems, err := oim.OrderItemRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list order items: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.OrderItemsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				orderItems,
			),
		)
	} else {
		_ = middleware.TemplRender(
			r,
			pages.OrderItemsPage(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				user.Username,
				ThemeToString(user.Theme),
				orderItems,
			),
		)
	}
}

// OrderItemPost creates a new order item.
func (oim *OrderItemManager) OrderItemPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	orderIDStr := strings.TrimSpace(r.Get("order_id").String())
	productDescription := strings.TrimSpace(r.Get("product_description").String())
	quantityStr := strings.TrimSpace(r.Get("quantity").String())
	unitPriceStr := strings.TrimSpace(r.Get("unit_price").String())
	totalPriceStr := strings.TrimSpace(r.Get("total_price").String())

	errs := map[string]string{}
	if orderIDStr == "" {
		errs["order_id"] = "Order ID is required"
	}

	var orderID int64
	if orderIDStr != "" {
		if idVal, err := strconv.ParseInt(orderIDStr, 10, 64); err == nil {
			orderID = idVal
		} else {
			errs["order_id"] = "Order ID must be a valid number"
		}
	}

	var productDesc *string
	if productDescription != "" {
		productDesc = &productDescription
	}

	var quantity *float64
	if quantityStr != "" {
		if qtyVal, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = &qtyVal
		} else {
			errs["quantity"] = "Quantity must be a valid number"
		}
	}

	var unitPrice *float64
	if unitPriceStr != "" {
		if priceVal, err := strconv.ParseFloat(unitPriceStr, 64); err == nil {
			unitPrice = &priceVal
		} else {
			errs["unit_price"] = "Unit price must be a valid number"
		}
	}

	var totalPrice *float64
	if totalPriceStr != "" {
		if priceVal, err := strconv.ParseFloat(totalPriceStr, 64); err == nil {
			totalPrice = &priceVal
		} else {
			errs["total_price"] = "Total price must be a valid number"
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		orderItem := &domain.OrderItem{
			OrderID:            orderID,
			ProductDescription: productDesc,
			Quantity:           quantity,
			UnitPrice:          unitPrice,
			TotalPrice:         totalPrice,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := oim.OrderItemRepo.Create(r.GetCtx(), orderItem)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create order item: %v", err)
			errs["form"] = "Failed to create order item"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/order-items")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/order-items")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/order-items")
}

// OrderItemGet renders a specific order item for editing or a new order item form.
func (oim *OrderItemManager) OrderItemGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var orderItem *domain.OrderItem

	// Check if this is a request for a new order item (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new order item
		orderItem = nil
	} else {
		// This is a request for editing an existing order item
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid order item ID")
			return
		}

		orderItem, err = oim.OrderItemRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Order item not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find order item: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	orders, err := oim.OrderRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list orders: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.OrderItemContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				orderItem,
				orders,
			),
		)
	} else {
		_ = middleware.TemplRender(
			r,
			pages.OrderItemPage(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				user.Username,
				ThemeToString(user.Theme),
				orderItem,
				orders,
			),
		)
	}
}

// OrderItemPut updates an existing order item.
func (oim *OrderItemManager) OrderItemPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid order item ID")
		return
	}

	orderIDStr := strings.TrimSpace(r.Get("order_id").String())
	productDescription := strings.TrimSpace(r.Get("product_description").String())
	quantityStr := strings.TrimSpace(r.Get("quantity").String())
	unitPriceStr := strings.TrimSpace(r.Get("unit_price").String())
	totalPriceStr := strings.TrimSpace(r.Get("total_price").String())

	errs := map[string]string{}
	if orderIDStr == "" {
		errs["order_id"] = "Order ID is required"
	}

	var orderID int64
	if orderIDStr != "" {
		if idVal, err := strconv.ParseInt(orderIDStr, 10, 64); err == nil {
			orderID = idVal
		} else {
			errs["order_id"] = "Order ID must be a valid number"
		}
	}

	var productDesc *string
	if productDescription != "" {
		productDesc = &productDescription
	}

	var quantity *float64
	if quantityStr != "" {
		if qtyVal, err := strconv.ParseFloat(quantityStr, 64); err == nil {
			quantity = &qtyVal
		} else {
			errs["quantity"] = "Quantity must be a valid number"
		}
	}

	var unitPrice *float64
	if unitPriceStr != "" {
		if priceVal, err := strconv.ParseFloat(unitPriceStr, 64); err == nil {
			unitPrice = &priceVal
		} else {
			errs["unit_price"] = "Unit price must be a valid number"
		}
	}

	var totalPrice *float64
	if totalPriceStr != "" {
		if priceVal, err := strconv.ParseFloat(totalPriceStr, 64); err == nil {
			totalPrice = &priceVal
		} else {
			errs["total_price"] = "Total price must be a valid number"
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		orderItem := &domain.OrderItem{
			OrderItemID:        id,
			OrderID:            orderID,
			ProductDescription: productDesc,
			Quantity:           quantity,
			UnitPrice:          unitPrice,
			TotalPrice:         totalPrice,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := oim.OrderItemRepo.Update(r.GetCtx(), orderItem)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update order item: %v", err)
			errs["form"] = "Failed to update order item"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/order-items/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/order-items/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated order item
	r.Response.RedirectTo(fmt.Sprintf("%s/management/order-items/%d", middleware.BasePath(), id))
}

// OrderItemDelete soft deletes an order item.
func (oim *OrderItemManager) OrderItemDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid order item ID")
		return
	}

	err = oim.OrderItemRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete order item: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/order-items")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/order-items")
}
