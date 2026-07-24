package service

import (
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestTopicService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	svc := NewTopicService(db)
	c := setupTestContext()

	// 1. Create
	topic := domain.Topic{
		TopicCode:     "TP001",
		TopicName:     "Dharma Discourse",
		TopicCategory: "LECTURE",
		Description:   "Weekly topic",
	}

	created, err := svc.Create(topic, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.TopicCode != "TP001" {
		t.Errorf("expected TP001, got %s", created.TopicCode)
	}

	// 2. Get
	fetched, err := svc.Get("TP001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.TopicName != "Dharma Discourse" {
		t.Errorf("expected 'Dharma Discourse', got '%s'", fetched.TopicName)
	}

	// 3. List
	filters := map[string]string{"topic_name": "Dharma"}
	items, total, err := svc.List(1, filters, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 item, got %d", total)
	}

	// 4. Update
	created.TopicName = "Advanced Dharma"
	updated, err := svc.Update("TP001", created, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.TopicName != "Advanced Dharma" {
		t.Errorf("expected Advanced Dharma, got %s", updated.TopicName)
	}

	// 5. Delete
	err = svc.Delete("TP001", c)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
