package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/omnsight/omndapi/gen/dapi/v1"
	"github.com/omnsight/omniscent-library/gen/model/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcHost    = "localhost"
	grpcPortEnv = "SERVICE_GRPC_PORT"
	defaultPort = "9091"
)

func getGRPCPort() string {
	port := os.Getenv(grpcPortEnv)
	if port == "" {
		return defaultPort
	}
	return port
}

func getMockToken() string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"preferred_username": "admin",
		"roles":              []string{"admin"},
		"exp":                time.Now().Add(time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	return fmt.Sprintf("%s.%s.mock-signature", headerEncoded, payloadEncoded)
}

func getAuthenticatedContext() context.Context {
	token := getMockToken()
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env.local")
	_ = godotenv.Load("../.env")
	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	// Connect to gRPC server
	target := fmt.Sprintf("%s:%s", grpcHost, getGRPCPort())
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	entityClient := dapi.NewEntityServiceClient(conn)
	relationClient := dapi.NewRelationshipServiceClient(conn)

	ctx := getAuthenticatedContext()

	t.Log("Using mock JWT token")

	// --- 3. Create Entities ---

	// Events
	event1 := &model.Event{
		Owner: "admin",
		Read:  []string{"admin"},
		Write: []string{"admin"},
		Title: "彩虹独角兽供应链魔法大会",
		Type:  "魔法大会",
		Location: &model.LocationData{
			Latitude:              36.8835,
			Longitude:             -123.43,
			CountryCode:           "FANTASY",
			AdministrativeArea:    "彩虹王国",
			SubAdministrativeArea: "魔法森林县",
			Locality:              "糖果城堡",
			SubLocality:           "彩虹大道",
			Address:               "123 彩虹独角兽街道",
			PostalCode:            99999,
		},
		Description: "本次大会由彩虹独角兽主办，讨论如何用魔法粉末重组供应链，使用彩虹桥运输货物，以及如何用魔法棒调整关税政策。与会者包括会说话的动物、会飞的汽车和会跳舞的树木，共同探讨用童话方式解决现实问题。",
		Tags:        []string{"产业", "供应链", "关税"},
		HappenedAt:  time.Now().Unix(),
	}
	respE1, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "event",
		Entity:     &model.Entity{Entity: &model.Entity_Event{Event: event1}},
	})
	if err != nil {
		t.Fatalf("Failed to create event1: %v", err)
	}
	e1 := respE1.Entity
	if e1.GetEvent().GetId() == "" {
		t.Error("e1 ID should be defined")
	}

	event2 := &model.Event{
		Owner: "admin",
		Read:  []string{"admin"},
		Write: []string{"admin"},
		Title: "星际穿越贸易博览会",
		Type:  "博览会",
		Location: &model.LocationData{
			Latitude:              40.7128,
			Longitude:             -74.0060,
			CountryCode:           "SPACE",
			AdministrativeArea:    "Milky Way",
			SubAdministrativeArea: "太阳系",
			Locality:              "火星殖民地",
			SubLocality:           "红色平原",
			Address:               "42 火星一大道",
			PostalCode:            0,
		},
		Description: "来自全宇宙的商人和探险家齐聚一堂，展示最新的光速飞船、量子通讯设备和反重力靴子。重点讨论如何建立跨星系自由贸易区，以及如何应对黑洞造成的物流延误。",
		Tags:        []string{"科技", "贸易", "星际"},
		HappenedAt:  time.Now().Unix(),
	}
	respE2, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "event",
		Entity:     &model.Entity{Entity: &model.Entity_Event{Event: event2}},
	})
	if err != nil {
		t.Fatalf("Failed to create event2: %v", err)
	}
	e2 := respE2.Entity

	event3 := &model.Event{
		Owner: "admin",
		Read:  []string{"admin"},
		Write: []string{"admin"},
		Title: "深海奇珍异宝拍卖会",
		Type:  "拍卖会",
		Location: &model.LocationData{
			Latitude:              -25.2744,
			Longitude:             133.7751,
			CountryCode:           "OCEAN",
			AdministrativeArea:    "太平洋海洋",
			SubAdministrativeArea: "马里亚纳 trench",
			Locality:              "亚特兰蒂斯废墟",
			SubLocality:           "珊瑚 plaza",
			Address:               "888 深海大道",
			PostalCode:            88888,
		},
		Description: "神秘的深海拍卖会，拍品包括美人鱼的眼泪、巨型章鱼的墨汁画和沉船中的古老金币。本次拍卖会所得款项将用于保护海洋生态环境，防止塑料垃圾污染海洋。",
		Tags:        []string{"拍卖", "珍宝", "环保"},
		HappenedAt:  time.Now().Unix(),
	}
	respE3, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "event",
		Entity:     &model.Entity{Entity: &model.Entity_Event{Event: event3}},
	})
	if err != nil {
		t.Fatalf("Failed to create event3: %v", err)
	}
	e3 := respE3.Entity

	// Persons
	person1 := &model.Person{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "甘道夫",
		Role:        "巫师",
		Nationality: "迈雅",
		BirthDate:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"巫师", "领袖", "魔法"},
		Aliases:     []string{"WG"},
	}
	respP1, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "person",
		Entity:     &model.Entity{Entity: &model.Entity_Person{Person: person1}},
	})
	if err != nil {
		t.Fatalf("Failed to create person1: %v", err)
	}
	p1 := respP1.Entity

	person2 := &model.Person{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "埃隆·马斯",
		Role:        "企业家",
		Nationality: "火星人",
		BirthDate:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"企业家", "科技", "太空"},
		Aliases:     []string{"EM"},
	}
	respP2, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "person",
		Entity:     &model.Entity{Entity: &model.Entity_Person{Person: person2}},
	})
	if err != nil {
		t.Fatalf("Failed to create person2: %v", err)
	}
	p2 := respP2.Entity

	person3 := &model.Person{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "尼莫船长",
		Role:        "船长",
		Nationality: "印度",
		BirthDate:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"船长", "探险家", "海洋"},
		Aliases:     []string{"Nemo"},
	}
	respP3, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "person",
		Entity:     &model.Entity{Entity: &model.Entity_Person{Person: person3}},
	})
	if err != nil {
		t.Fatalf("Failed to create person3: %v", err)
	}
	p3 := respP3.Entity

	// Organizations
	org1 := &model.Organization{
		Owner:        "admin",
		Read:         []string{"admin"},
		Write:        []string{"admin"},
		Name:         "独角兽供应链公司",
		Type:         "盈利性公司",
		FoundedAt:    time.Now().Unix(),
		DiscoveredAt: time.Now().Unix(),
		LastVisited:  time.Now().Unix(),
		Tags:         []string{"物流", "魔法", "供应链"},
	}
	respO1, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "organization",
		Entity:     &model.Entity{Entity: &model.Entity_Organization{Organization: org1}},
	})
	if err != nil {
		t.Fatalf("Failed to create org1: %v", err)
	}
	o1 := respO1.Entity

	org2 := &model.Organization{
		Owner:        "admin",
		Read:         []string{"admin"},
		Write:        []string{"admin"},
		Name:         "星际探索者",
		Type:         "盈利性公司",
		FoundedAt:    time.Now().Unix(),
		DiscoveredAt: time.Now().Unix(),
		LastVisited:  time.Now().Unix(),
		Tags:         []string{"航天", "科技", "探索"},
	}
	respO2, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "organization",
		Entity:     &model.Entity{Entity: &model.Entity_Organization{Organization: org2}},
	})
	if err != nil {
		t.Fatalf("Failed to create org2: %v", err)
	}
	o2 := respO2.Entity

	org3 := &model.Organization{
		Owner:        "admin",
		Read:         []string{"admin"},
		Write:        []string{"admin"},
		Name:         "深蓝保护协会",
		Type:         "非营利组织",
		FoundedAt:    time.Now().Unix(),
		DiscoveredAt: time.Now().Unix(),
		LastVisited:  time.Now().Unix(),
		Tags:         []string{"环保", "海洋", "公益"},
	}
	respO3, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "organization",
		Entity:     &model.Entity{Entity: &model.Entity_Organization{Organization: org3}},
	})
	if err != nil {
		t.Fatalf("Failed to create org3: %v", err)
	}
	o3 := respO3.Entity

	// Websites
	web1 := &model.Website{
		Owner:        "admin",
		Read:         []string{"admin"},
		Write:        []string{"admin"},
		Url:          fmt.Sprintf("https://www.magic-supply-chain-%d.fantasy", time.Now().UnixNano()),
		Title:        "独角兽供应链官方网",
		Description:  "独角兽供应链公司的官方网站，涵盖从供应链物流到魔法业务的所有内容。",
		FoundedAt:    time.Now().Unix(),
		DiscoveredAt: time.Now().Unix(),
		LastVisited:  time.Now().Unix(),
		Tags:         []string{"魔法", "商业"},
	}
	respW1, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "website",
		Entity:     &model.Entity{Entity: &model.Entity_Website{Website: web1}},
	})
	if err != nil {
		t.Fatalf("Failed to create web1: %v", err)
	}
	w1 := respW1.Entity

	web2 := &model.Website{
		Owner:        "admin",
		Read:         []string{"admin"},
		Write:        []string{"admin"},
		Url:          fmt.Sprintf("https://www.mars-colonies-%d.space", time.Now().UnixNano()),
		Title:        "火星殖民地快讯",
		Description:  "关于火星殖民地建设、生活和贸易的最新消息。",
		FoundedAt:    time.Now().Unix(),
		DiscoveredAt: time.Now().Unix(),
		LastVisited:  time.Now().Unix(),
		Tags:         []string{"太空", "新闻"},
	}
	respW2, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "website",
		Entity:     &model.Entity{Entity: &model.Entity_Website{Website: web2}},
	})
	if err != nil {
		t.Fatalf("Failed to create web2: %v", err)
	}
	w2 := respW2.Entity

	// Sources
	src1 := &model.Source{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "预言家日报",
		Type:        "新闻",
		Url:         "https://www.thedailyprophet.com/article-one",
		Reliability: 20,
		Title:       "头条新闻：魔法供应链的崛起",
		Description: "预言家日报发布了关于魔法供应链的最新消息，这将改变整个魔法行业的运营模式。",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"新闻", "魔法"},
	}
	respS1, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "source",
		Entity:     &model.Entity{Entity: &model.Entity_Source{Source: src1}},
	})
	if err != nil {
		t.Fatalf("Failed to create src1: %v", err)
	}
	s1 := respS1.Entity

	src2 := &model.Source{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "银河公报",
		Type:        "新闻",
		Url:         "https://www.galactic-gazette.space/daily-news",
		Reliability: 80,
		Title:       "银河系贸易协定签署",
		Description: "银河公报发布了关于银河系贸易协定签署的最新消息，这将改变整个宇宙的贸易模式。",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"新闻", "太空"},
	}
	respS2, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "source",
		Entity:     &model.Entity{Entity: &model.Entity_Source{Source: src2}},
	})
	if err != nil {
		t.Fatalf("Failed to create src2: %v", err)
	}
	s2 := respS2.Entity

	src3 := &model.Source{
		Owner:       "admin",
		Read:        []string{"admin"},
		Write:       []string{"admin"},
		Name:        "深海回声",
		Type:        "新闻",
		Url:         "https://www.deep-sea-echo.ocean/reports",
		Reliability: 90,
		Title:       "亚特兰蒂斯遗址的新发现",
		Description: "深海回声团队发现了亚特兰蒂斯遗址的新发现，这是一个神秘的岛屿，包含了古代的魔法力量。",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Tags:        []string{"新闻", "海洋"},
	}
	respS3, err := entityClient.CreateEntity(ctx, &dapi.CreateEntityRequest{
		EntityType: "source",
		Entity:     &model.Entity{Entity: &model.Entity_Source{Source: src3}},
	})
	if err != nil {
		t.Fatalf("Failed to create src3: %v", err)
	}
	s3 := respS3.Entity

	// --- 4. Connect Entities ---
	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    p1.GetPerson().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "participant",
			Label: "参与者",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->p1: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    p2.GetPerson().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "hosted_by",
			Label: "主持人",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->p2: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    o1.GetOrganization().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "organized_by",
			Label: "主办方",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->o1: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    w1.GetWebsite().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "mentioned_by",
			Label: "提及网站",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->w1: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    s1.GetSource().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "reported_by",
			Label: "报道信源",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->s1: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    s2.GetSource().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "sourced_from",
			Label: "信源",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->s2: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e2.GetEvent().GetId(),
			To:    p2.GetPerson().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "sponsor",
			Label: "赞助商",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e2->p2: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e2.GetEvent().GetId(),
			To:    o2.GetOrganization().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "sponsor",
			Label: "赞助商",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e2->o2: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e2.GetEvent().GetId(),
			To:    s2.GetSource().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "sourced_from",
			Label: "信源",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e2->s2: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  o1.GetOrganization().GetId(),
			To:    w1.GetWebsite().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "has_website",
			Label: "公司网站",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship o1->w1: %v", err)
	}

	_, err = relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e1.GetEvent().GetId(),
			To:    e2.GetEvent().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "related_event",
			Label: "相关事件",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create relationship e1->e2: %v", err)
	}

	respRel, err := relationClient.CreateRelationship(ctx, &dapi.CreateRelationshipRequest{
		Relationship: &model.Relation{
			From:  e3.GetEvent().GetId(),
			To:    p3.GetPerson().GetId(),
			Owner: "admin",
			Read:  []string{"admin"},
			Write: []string{"admin"},
			Name:  "temp_relation",
			Label: "临时关系",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create temp relationship: %v", err)
	}
	tempRel := respRel.Relationship

	// --- 4.5 List Entities from Event ---
	nowTime := time.Now()
	startOfDay := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location()).Unix()
	endOfDay := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 23, 59, 59, 0, nowTime.Location()).Unix()

	// Test with startNode and time range
	list1, err := entityClient.ListEntitiesFromEvent(ctx, &dapi.ListEntitiesFromEventRequest{
		StartNode: e1.GetEvent().GetId(),
		Depth:     1,
	})
	if err != nil {
		t.Fatalf("Failed to list entities from event (1): %v", err)
	}
	t.Logf("ListEntitiesFromEvent 1: %d entities, %d relations", len(list1.Entities), len(list1.Relations))
	if len(list1.Entities) == 0 {
		t.Fatal("ListEntitiesFromEvent 1 returned 0 entities")
	}

	// Test with other query parameters
	list2, err := entityClient.ListEntitiesFromEvent(ctx, &dapi.ListEntitiesFromEventRequest{
		StartNode:   e1.GetEvent().GetId(),
		CountryCode: "FANTASY",
		Tag:         "产业",
		Depth:       1,
	})
	if err != nil {
		t.Fatalf("Failed to list entities from event (2): %v", err)
	}
	t.Logf("ListEntitiesFromEvent 2: %d entities, %d relations", len(list2.Entities), len(list2.Relations))
	if len(list2.Entities) == 0 || len(list2.Relations) == 0 {
		t.Fatal("ListEntitiesFromEvent 2 (filtered) returned 0 entities or relations")
	}

	// Test with startNode and time range
	list3, err := entityClient.ListEntitiesFromEvent(ctx, &dapi.ListEntitiesFromEventRequest{
		StartTime: startOfDay,
		EndTime:   endOfDay,
	})
	if err != nil {
		t.Fatalf("Failed to list entities in time range: %v", err)
	}
	t.Logf("ListEntitiesFromEvent 3: %d entities, %d relations", len(list3.Entities), len(list3.Relations))
	if len(list3.Entities) == 0 || len(list3.Relations) == 0 {
		t.Fatal("ListEntitiesFromEvent 3 returned 0 entities or relations")
	}

	// --- 5. Delete One of Each ---

	// Delete Temp Relation
	_, err = relationClient.DeleteRelationship(ctx, &dapi.DeleteRelationshipRequest{
		Collection: "event_temp_relation_person",
		Key:        tempRel.GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete temp relationship: %v", err)
	}

	// Delete Source 3
	_, err = entityClient.DeleteEntity(ctx, &dapi.DeleteEntityRequest{
		EntityType: "source",
		Key:        s3.GetSource().GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete s3: %v", err)
	}

	// Delete Website 2
	_, err = entityClient.DeleteEntity(ctx, &dapi.DeleteEntityRequest{
		EntityType: "website",
		Key:        w2.GetWebsite().GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete w2: %v", err)
	}

	// Delete Organization 3
	_, err = entityClient.DeleteEntity(ctx, &dapi.DeleteEntityRequest{
		EntityType: "organization",
		Key:        o3.GetOrganization().GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete o3: %v", err)
	}

	// Delete Person 3
	_, err = entityClient.DeleteEntity(ctx, &dapi.DeleteEntityRequest{
		EntityType: "person",
		Key:        p3.GetPerson().GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete p3: %v", err)
	}

	// Delete Event 3
	_, err = entityClient.DeleteEntity(ctx, &dapi.DeleteEntityRequest{
		EntityType: "event",
		Key:        e3.GetEvent().GetKey(),
	})
	if err != nil {
		t.Fatalf("Failed to delete e3: %v", err)
	}

	t.Log("Integration test finished successfully")
}
