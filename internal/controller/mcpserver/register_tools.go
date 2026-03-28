package mcpserver

func (h *Handler) registerTools() {
	h.registerMathTools()
	h.registerTimeTools()
	h.registerClickHouseTools()
}
