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

type OrderManager struct {
	OrderRepo    data.OrderRepo
	CustomerRepo data.CustomerRepo
}

// RegisterOrderRoutes wires order management endpoints under /app.
func RegisterOrderRoutes(group *ghttp.RouterGroup, orderRepo data.OrderRepo, customerRepo data.CustomerRepo) {
	om := &OrderManager{
		OrderRepo:    orderRepo,
		CustomerRepo: customerRepo,
	}

	// Order management
	group.GET("/management/orders", om.OrdersGet)
	group.POST("/management/orders", om.OrderPost)
	group.GET("/management/orders/new", om.OrderGet)
	group.GET("/management/orders/:id", om.OrderGet)
	group.PUT("/management/orders/:id", om.OrderPut)
	group.DELETE("/management/orders/:id", om.OrderDelete)
}

// OrdersGet renders the orders management page.
func (om *OrderManager) OrdersGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	orders, err := om.OrderRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list orders: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.OrdersContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				orders,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.OrdersPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			orders,
		),
	)
}

// OrderPost creates a new order.
func (om *OrderManager) OrderPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	customerIDStr := strings.TrimSpace(r.Get("customer_id").String())
	orderDateStr := strings.TrimSpace(r.Get("order_date").String())
	deliveryDateStr := strings.TrimSpace(r.Get("delivery_date").String())
	totalAmountStr := strings.TrimSpace(r.Get("total_amount").String())
	status := strings.TrimSpace(r.Get("status").String())

	errs := map[string]string{}
	if customerIDStr == "" {
		errs["customer_id"] = "Customer ID is required"
	}

	var customerID int64
	if customerIDStr != "" {
		if idVal, err := strconv.ParseInt(customerIDStr, 10, 64); err == nil {
			customerID = idVal
		} else {
			errs["customer_id"] = "Customer ID must be a valid number"
		}
	}

	var orderDate *time.Time
	if orderDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", orderDateStr); err == nil {
			orderDate = new(time.Time)
			*orderDate = parsedDate
		} else {
			errs["order_date"] = "Order date must be in YYYY-MM-DD format"
		}
	}

	var deliveryDate *time.Time
	if deliveryDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", deliveryDateStr); err == nil {
			deliveryDate = new(time.Time)
			*deliveryDate = parsedDate
		} else {
			errs["delivery_date"] = "Delivery date must be in YYYY-MM-DD format"
		}
	}

	var totalAmount *float64
	if totalAmountStr != "" {
		if amountVal, err := strconv.ParseFloat(totalAmountStr, 64); err == nil {
			totalAmount = new(float64)
			*totalAmount = amountVal
		} else {
			errs["total_amount"] = "Total amount must be a valid number"
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = new(string)
		*statusPtr = status
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		order := &domain.Order{
			CustomerID:   customerID,
			OrderDate:    orderDate,
			DeliveryDate: deliveryDate,
			TotalAmount:  totalAmount,
			Status:       statusPtr,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := om.OrderRepo.Create(r.GetCtx(), order)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create order: %v", err)
			errs["form"] = "Failed to create order"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/orders")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/orders")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/orders")
}

// OrderGet renders a specific order for editing or a new order form.
func (om *OrderManager) OrderGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var order *domain.Order

	// Check if this is a request for a new order (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new order
		order = nil
	} else {
		// This is a request for editing an existing order
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid order ID")
			return
		}

		order, err = om.OrderRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Order not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find order: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	// Fetch customers for dropdown
	customers, err := om.CustomerRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list customers: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.OrderContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				order,
				customers,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.OrderPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			order,
			customers,
		),
	)
}

// OrderPut updates an existing order.
func (om *OrderManager) OrderPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid order ID")
		return
	}

	customerIDStr := strings.TrimSpace(r.Get("customer_id").String())
	orderDateStr := strings.TrimSpace(r.Get("order_date").String())
	deliveryDateStr := strings.TrimSpace(r.Get("delivery_date").String())
	totalAmountStr := strings.TrimSpace(r.Get("total_amount").String())
	status := strings.TrimSpace(r.Get("status").String())

	errs := map[string]string{}
	if customerIDStr == "" {
		errs["customer_id"] = "Customer ID is required"
	}

	var customerID int64
	if customerIDStr != "" {
		if idVal, err := strconv.ParseInt(customerIDStr, 10, 64); err == nil {
			customerID = idVal
		} else {
			errs["customer_id"] = "Customer ID must be a valid number"
		}
	}

	var orderDate *time.Time
	if orderDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", orderDateStr); err == nil {
			orderDate = new(time.Time)
			*orderDate = parsedDate
		} else {
			errs["order_date"] = "Order date must be in YYYY-MM-DD format"
		}
	}

	var deliveryDate *time.Time
	if deliveryDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", deliveryDateStr); err == nil {
			deliveryDate = new(time.Time)
			*deliveryDate = parsedDate
		} else {
			errs["delivery_date"] = "Delivery date must be in YYYY-MM-DD format"
		}
	}

	var totalAmount *float64
	if totalAmountStr != "" {
		if amountVal, err := strconv.ParseFloat(totalAmountStr, 64); err == nil {
			totalAmount = new(float64)
			*totalAmount = amountVal
		} else {
			errs["total_amount"] = "Total amount must be a valid number"
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = new(string)
		*statusPtr = status
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		order := &domain.Order{
			OrderID:      id,
			CustomerID:   customerID,
			OrderDate:    orderDate,
			DeliveryDate: deliveryDate,
			TotalAmount:  totalAmount,
			Status:       statusPtr,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := om.OrderRepo.Update(r.GetCtx(), order)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update order: %v", err)
			errs["form"] = "Failed to update order"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/orders/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/orders/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated order
	r.Response.RedirectTo(fmt.Sprintf("%s/management/orders/%d", middleware.BasePath(), id))
}

// OrderDelete soft deletes an order.
func (om *OrderManager) OrderDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid order ID")
		return
	}

	err = om.OrderRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete order: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/orders")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/orders")
}
