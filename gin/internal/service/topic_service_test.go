package service

import (
	"errors"
	"testing"

	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestTopicService(t *testing.T) {
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	svc := NewTopicService(db)
	c := setupTestContext()

	db.Exec("DELETE FROM T_BUS_TOPIC WHERE TopicCode IN ('TP001', 'TP002')")

	// 1. Create
	topic := domain.Topic{
		TopicCode:     "TP001",
		TopicName:     "Dharma Discourse",
		TopicCategory: "LEC",
		Description:   "Weekly topic",
	}

	created, err := svc.Create(topic, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.TopicCode != "TP001" {
		t.Errorf("expected TP001, got %s", created.TopicCode)
	}

	// 1b. Create Duplicate Name (should fail)
	dupTopic := domain.Topic{
		TopicCode:     "TP002",
		TopicName:     "Dharma Discourse",
		TopicCategory: "LEC",
		Description:   "Duplicate name topic",
	}
	_, err = svc.Create(dupTopic, c)
	if err == nil {
		t.Fatalf("expected error when creating topic with existing topic_name, got nil")
	}
	t.Logf(">>> 1b error is: %v (type: %T)", err, err)

	// 1c. Create second topic with different name
	dupTopic.TopicName = "Second Topic"
	topic2, err := svc.Create(dupTopic, c)
	if err != nil {
		t.Fatalf("Create second topic failed: %v", err)
	}

	// 1d. Create with BOTH struct validation error (TopicCategory max=3) and duplicate topic_name
	invalidAndDupTopic := domain.Topic{
		TopicCode:     "TP003",
		TopicName:     "Dharma Discourse", // duplicate of TP001
		TopicCategory: "TOOLONGVALUE",     // exceeds max=3
	}
	_, err = svc.Create(invalidAndDupTopic, c)
	if err == nil {
		t.Fatalf("expected combined validation error, got nil")
	}
	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if len(vErr.Details["topic_name"]) == 0 {
		t.Errorf("expected topic_name error in details, got: %v", vErr.Details)
	}
	if len(vErr.Details["topic_category"]) == 0 {
		t.Errorf("expected topic_category error in details, got: %v", vErr.Details)
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

	// 4a. Update keeping the SAME topic_name (should succeed)
	created.Description = "Updated description only"
	updatedSameName, err := svc.Update("TP001", created, c)
	if err != nil {
		t.Fatalf("Update with unchanged topic_name should succeed, got: %v", err)
	}
	if updatedSameName.Description != "Updated description only" {
		t.Errorf("expected updated description, got %s", updatedSameName.Description)
	}

	// 4b. Update changing to another existing topic's name (should fail)
	created.TopicName = "Second Topic"
	_, err = svc.Update("TP001", created, c)
	if err == nil {
		t.Fatalf("expected error when updating topic to another existing topic_name, got nil")
	}

	// 4c. Update with BOTH struct validation error and duplicate topic_name
	invalidUpdate := created
	invalidUpdate.TopicName = "Second Topic" // duplicate of topic2
	invalidUpdate.TopicCategory = "TOOLONGVALUE"
	_, err = svc.Update("TP001", invalidUpdate, c)
	if err == nil {
		t.Fatalf("expected combined validation error on update, got nil")
	}
	var vErrUpdate *ValidationError
	if !errors.As(err, &vErrUpdate) {
		t.Fatalf("expected *ValidationError on update, got %T: %v", err, err)
	}
	if len(vErrUpdate.Details["topic_name"]) == 0 {
		t.Errorf("expected topic_name error in update details, got: %v", vErrUpdate.Details)
	}
	if len(vErrUpdate.Details["topic_category"]) == 0 {
		t.Errorf("expected topic_category error in update details, got: %v", vErrUpdate.Details)
	}

	// 4d. Update to a new unique topic_name (should succeed)
	created.TopicName = "Advanced Dharma"
	created.TopicCategory = "LEC"
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
	err = svc.Delete(topic2.TopicCode, c)
	if err != nil {
		t.Fatalf("Delete topic2 failed: %v", err)
	}
}
