package server

import "github.com/gin-gonic/gin"

func Run(r *gin.Engine, addr string) error {
	return r.Run(addr)
}
