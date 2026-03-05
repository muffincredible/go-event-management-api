package middleware

import (
	"os"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Yetkisiz giriş, token bulunamadı"})
	}

	tokenString := strings.Split(authHeader, "Bearer ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Geçersiz veya süresi dolmuş token"})
	}

	claims := token.Claims.(jwt.MapClaims)
	c.Locals("userId", claims["id"])

	return c.Next()
}