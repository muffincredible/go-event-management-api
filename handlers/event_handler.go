package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/muffincredible/go-event-management-api/configs"
	"github.com/muffincredible/go-event-management-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var eventCollection *mongo.Collection = configs.GetCollection(configs.ConnectDB(), "events")

//etkinlik oluşturma
func CreateEvent(c *fiber.Ctx) error {
	var event models.Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz veri"})
	}

	//geçmiş tarihli etkinlik oluşturulamaz
	if event.Date.Before(time.Now()) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Etkinlik tarihi geçmiş bir zaman olamaz"})
	}

	userIdStr := c.Locals("userId").(string)
	objId, _ := primitive.ObjectIDFromHex(userIdStr)
	event.CreatorId = objId
	event.Participants = []primitive.ObjectID{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := eventCollection.InsertOne(ctx, event)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Etkinlik oluşturulamadı"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": "Etkinlik başarıyla oluşturuldu"})
}

//etkinliğe katılma
func JoinEvent(c *fiber.Ctx) error {
	eventId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	userIdStr := c.Locals("userId").(string)
	userId, _ := primitive.ObjectIDFromHex(userIdStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event models.Event
	err := eventCollection.FindOne(ctx, bson.M{"_id": eventId}).Decode(&event)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Etkinlik bulunamadı"})
	}

	//kapasite kontrolü
	if len(event.Participants) >= event.Capacity {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Etkinlik kapasitesi dolu"})
	}

	//aynı etkinliğe 1den fazla katılım engelleme
	for _, p := range event.Participants {
		if p == userId {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Zaten bu etkinliğe katılmışsınız"})
		}
	}

	_, err = eventCollection.UpdateOne(ctx, bson.M{"_id": eventId}, bson.M{"$push": bson.M{"participants": userId}})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Katılım işlemi başarısız"})
	}

	return c.JSON(fiber.Map{"message": "Etkinliğe başarıyla katıldınız"})
}

//etkinlik silme
func DeleteEvent(c *fiber.Ctx) error {
	eventId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	userIdStr := c.Locals("userId").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event models.Event
	err := eventCollection.FindOne(ctx, bson.M{"_id": eventId}).Decode(&event)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Etkinlik bulunamadı"})
	}

	if event.CreatorId.Hex() != userIdStr {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Bu işlemi yapmaya yetkiniz yok"})
	}

	_, err = eventCollection.DeleteOne(ctx, bson.M{"_id": eventId})
	return c.JSON(fiber.Map{"message": "Etkinlik silindi"})
}

//tüm etkinlikleri listele
func ListEvents(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := eventCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Veriler alınamadı"})
	}

	var events []models.Event
	cursor.All(ctx, &events)
	return c.JSON(events)
}

//etkinlik güncelleme
func UpdateEvent(c *fiber.Ctx) error {
	eventId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	userIdStr := c.Locals("userId").(string)

	var updateData models.Event
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz veri formatı"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event models.Event
	err := eventCollection.FindOne(ctx, bson.M{"_id": eventId}).Decode(&event)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Etkinlik bulunamadı"})
	}

	if event.CreatorId.Hex() != userIdStr {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "Bu etkinliği güncelleme yetkiniz yok"})
	}

	update := bson.M{
		"$set": bson.M{
			"title":       updateData.Title,
			"description": updateData.Description,
			"date":        updateData.Date,
			"capacity":    updateData.Capacity,
			"location":    updateData.Location,
		},
	}

	_, err = eventCollection.UpdateOne(ctx, bson.M{"_id": eventId}, update)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Güncelleme başarısız"})
	}

	return c.JSON(fiber.Map{"message": "Etkinlik başarıyla güncellendi"})
}

func LeaveEvent(c *fiber.Ctx) error {
	eventId, _ := primitive.ObjectIDFromHex(c.Params("id"))
	userIdStr := c.Locals("userId").(string)
	userId, _ := primitive.ObjectIDFromHex(userIdStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := eventCollection.UpdateOne(ctx, bson.M{"_id": eventId}, bson.M{"$pull": bson.M{"participants": userId}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ayrılma işlemi başarısız"})
	}

	return c.JSON(fiber.Map{"message": "Etkinlikten başarıyla ayrıldınız"})
}

//katıldığınız etkinlikleri listeleme
func GetMyEvents(c *fiber.Ctx) error {
	userIdStr := c.Locals("userId").(string)
	userId, _ := primitive.ObjectIDFromHex(userIdStr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := eventCollection.Find(ctx, bson.M{"participants": userId})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Veriler getirilemedi"})
	}

	var events []models.Event
	cursor.All(ctx, &events)
	return c.JSON(events)
}