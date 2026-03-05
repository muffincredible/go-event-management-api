package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/muffincredible/go-event-management-api/configs"
	"github.com/muffincredible/go-event-management-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.ConnectDB(), "users")

//register
func Register(c *fiber.Ctx) error {
	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz veri formatı"})
	}

	//şifre hashleme
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Şifre işlenirken bir hata oluştu"})
	}
	user.Password = string(hashedPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//kullanıcıyı veritabanına ekleme
	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Kullanıcı oluşturulamadı"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Kayıt başarılı!"})
}

//login
func Login(c *fiber.Ctx) error {
	var request models.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek formatı"})
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := userCollection.FindOne(ctx, bson.M{"email": request.Email}).Decode(&user)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "E-posta veya şifre hatalı"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "E-posta veya şifre hatalı"})
	}

	//JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.Id.Hex(),
		"exp": time.Now().Add(time.Hour * 72).Unix(), // 3 günlük geçerlilik
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Token oluşturulamadı"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Giriş başarılı",
		"token":   tokenString,
		"user":    user.Username,
	})
}