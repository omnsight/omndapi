import {
  Api,
  V1Entity,
} from './api/Api';

const CLIENT_ID = 'omndapi';
const API_BASE_URL = 'http://localhost:8081';

describe('Integration Test', () => {
  let accessToken: string;
  let api: Api<string>;

  beforeAll(async () => {
    // Mock JWT token generation
    const header = { alg: 'HS256', typ: 'JWT' };
    const payload = {
      sub: 'admin',
      resource_access: {
        [CLIENT_ID]: {
          roles: ['admin'],
        },
      },
      exp: Math.floor(Date.now() / 1000) + 3600,
    };

    const base64UrlEncode = (obj: any) => {
      return Buffer.from(JSON.stringify(obj)).toString('base64')
        .replace(/=/g, '')
        .replace(/\+/g, '-')
        .replace(/\//g, '_');
    };

    accessToken = `${base64UrlEncode(header)}.${base64UrlEncode(payload)}.mock-signature`;

    api = new Api({
      baseURL: API_BASE_URL,
      securityWorker: (token) => token ? { headers: { Authorization: `Bearer ${token}` } } : {},
      secure: true,
    });
    api.setSecurityData(accessToken);

    console.log('Using mock JWT token');
  });

  it('should create entities, relationships and delete them', async () => {
    // --- 3. Create Entities ---

    // Events
    const event1: V1Entity = {
      event: {
        title: "彩虹独角兽供应链魔法大会",
        roles: ["admin"],
        location: {
          latitude: 36.8835,
          longitude: -123.43,
          countryCode: "FANTASY",
          administrativeArea: "彩虹王国",
          subAdministrativeArea: "魔法森林县",
          locality: "糖果城堡",
          subLocality: "彩虹大道",
          address: "123 彩虹独角兽大街",
          postalCode: 99999,
        },
        description: "本次大会由彩虹独角兽主办，讨论如何用魔法粉末重组供应链，使用彩虹桥运输货物，以及如何用魔法棒调整关税政策。与会者包括会说话的动物、会飞的汽车和会跳舞的树木，共同探讨用童话方式解决现实问题。",
        tags: ["产业", "供应链", "关税"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
      }
    };
    const { data: respE1 } = await api.v1.entityServiceCreateEntity("event", event1);
    const e1 = respE1.entity;
    expect(e1?.event?.id).toBeDefined();

    const event2: V1Entity = {
      event: {
        title: "星际穿越贸易博览会",
        roles: ["admin"],
        location: {
          latitude: 40.7128,
          longitude: -74.0060,
          countryCode: "SPACE",
          administrativeArea: "银河系",
          subAdministrativeArea: "太阳系",
          locality: "火星殖民地",
          subLocality: "红色平原",
          address: "42 火星一号大道",
          postalCode: 0,
        },
        description: "来自全宇宙的商人和探险家齐聚一堂，展示最新的光速飞船、量子通讯设备和反重力靴子。重点讨论如何建立跨星系自由贸易区，以及如何应对黑洞造成的物流延误。",
        tags: ["科技", "贸易", "星际"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
      }
    };
    const { data: respE2 } = await api.v1.entityServiceCreateEntity("event", event2);
    const e2 = respE2.entity;

    const event3: V1Entity = {
      event: {
        title: "深海奇珍异宝拍卖会",
        roles: ["admin"],
        location: {
          latitude: -25.2744,
          longitude: 133.7751,
          countryCode: "OCEAN",
          administrativeArea: "太平洋",
          subAdministrativeArea: "马里亚纳海沟",
          locality: "亚特兰蒂斯遗址",
          subLocality: "珊瑚广场",
          address: "888 深海大道",
          postalCode: 88888,
        },
        description: "神秘的深海拍卖会，拍品包括美人鱼的眼泪、巨型章鱼的墨汁画和沉船中的古老金币。本次拍卖会所得款项将用于保护海洋生态环境，防止塑料垃圾污染海洋。",
        tags: ["拍卖", "珍宝", "环保"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
      }
    };
    const { data: respE3 } = await api.v1.entityServiceCreateEntity("event", event3);
    const e3 = respE3.entity;

    // Persons
    const person1: V1Entity = {
      person: {
        roles: ["admin"],
        name: "白袍甘道夫",
        role: "巫师",
        nationality: "迈雅",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["巫师", "领袖", "魔法"],
        aliases: ["WG"],
      }
    };
    const { data: respP1 } = await api.v1.entityServiceCreateEntity("person", person1);
    const p1 = respP1.entity;

    const person2: V1Entity = {
      person: {
        roles: ["admin"],
        name: "埃隆·马斯",
        role: "企业家",
        nationality: "火星人",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["企业家", "科技", "太空"],
        aliases: ["EM"],
      }
    };
    const { data: respP2 } = await api.v1.entityServiceCreateEntity("person", person2);
    const p2 = respP2.entity;

    const person3: V1Entity = {
      person: {
        roles: ["admin"],
        name: "尼莫船长",
        role: "船长",
        nationality: "印度",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["船长", "探险家", "海洋"],
        aliases: ["Nemo"],
      }
    };
    const { data: respP3 } = await api.v1.entityServiceCreateEntity("person", person3);
    const p3 = respP3.entity;

    // Organizations
    const org1: V1Entity = {
      organization: {
        roles: ["admin"],
        name: "独角兽供应链公司",
        type: "盈利性公司",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["物流", "魔法", "供应链"],
      }
    };
    const { data: respO1 } = await api.v1.entityServiceCreateEntity("organization", org1);
    const o1 = respO1.entity;

    const org2: V1Entity = {
      organization: {
        roles: ["admin"],
        name: "星际探索者",
        type: "盈利性公司",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["航天", "科技", "探索"],
      }
    };
    const { data: respO2 } = await api.v1.entityServiceCreateEntity("organization", org2);
    const o2 = respO2.entity;

    const org3: V1Entity = {
      organization: {
        roles: ["admin"],
        name: "深蓝保护协会",
        type: "非营利组织",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["环保", "海洋", "公益"],
      }
    };
    const { data: respO3 } = await api.v1.entityServiceCreateEntity("organization", org3);
    const o3 = respO3.entity;

    // Websites
    const web1: V1Entity = {
      website: {
        roles: ["admin"],
        url: "https://www.magic-supply-chain.fantasy",
        domain: "magic-supply-chain.fantasy",
        title: "独角兽供应链官方网",
        description: "独角兽供应链公司的官方网站，涵盖从供应链物流到魔法业务的所有内容。",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["魔法", "商业"],
      }
    };
    const { data: respW1 } = await api.v1.entityServiceCreateEntity("website", web1);
    const w1 = respW1.entity;

    const web2: V1Entity = {
      website: {
        roles: ["admin"],
        url: "https://www.mars-colonies.space",
        domain: "mars-colonies.space",
        title: "火星殖民地快讯",
        description: "关于火星殖民地建设、生活和贸易的最新消息。",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["太空", "新闻"],
      }
    };
    const { data: respW2 } = await api.v1.entityServiceCreateEntity("website", web2);
    const w2 = respW2.entity;

    // Sources
    const src1: V1Entity = {
      source: {
        roles: ["admin"],
        name: "预言家日报",
        url: "https://www.thedailyprophet.com/article-one",
        rootUrl: "https://www.thedailyprophet.com",
        reliability: 20,
        title: "头条新闻：魔法供应链的崛起",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["新闻", "魔法"],
      }
    };
    const { data: respS1 } = await api.v1.entityServiceCreateEntity("source", src1);
    const s1 = respS1.entity;

    const src2: V1Entity = {
      source: {
        roles: ["admin"],
        name: "银河公报",
        url: "https://www.galactic-gazette.space/daily-news",
        rootUrl: "https://www.galactic-gazette.space",
        reliability: 80,
        title: "银河系贸易协定签署",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["新闻", "太空"],
      }
    };
    const { data: respS2 } = await api.v1.entityServiceCreateEntity("source", src2);
    const s2 = respS2.entity;

    const src3: V1Entity = {
      source: {
        roles: ["admin"],
        name: "深海回声",
        url: "https://www.deep-sea-echo.ocean/reports",
        rootUrl: "https://www.deep-sea-echo.ocean",
        reliability: 90,
        title: "亚特兰蒂斯遗址的新发现",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["新闻", "海洋"],
      }
    };
    const { data: respS3 } = await api.v1.entityServiceCreateEntity("source", src3);
    const s3 = respS3.entity;

    // --- 4. Connect Entities ---
    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: p1?.person?.id,
      name: "participant",
      attributes: {
        zh: {
          name: "参与者",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: p2?.person?.id,
      name: "hosted_by",
      attributes: {
        zh: {
          name: "主持人",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: o1?.organization?.id,
      name: "organized_by",
      attributes: {
        zh: {
          name: "主办方",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: s1?.source?.id,
      name: "reported_by",
      attributes: {
        zh: {
          name: "报道信源",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: s2?.source?.id,
      name: "sourced_from",
      attributes: {
        zh: {
          name: "信源",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e2?.event?.id,
      to: p2?.person?.id,
      name: "sponsor",
      attributes: {
        zh: {
          name: "赞助者",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e2?.event?.id,
      to: o2?.organization?.id,
      name: "sponsor",
      attributes: {
        zh: {
          name: "赞助商",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: e2?.event?.id,
      to: s2?.source?.id,
      name: "sourced_from",
      attributes: {
        zh: {
          name: "信源",
        },
      },
    });

    await api.v1.relationshipServiceCreateRelationship({
      from: o1?.organization?.id,
      to: w1?.website?.id,
      name: "has_website",
      attributes: {
        zh: {
          name: "公司网站",
        },
      },
    });

    const { data: respRel } = await api.v1.relationshipServiceCreateRelationship({
      from: e3?.event?.id,
      to: p3?.person?.id,
      name: "temp_relation",
      attributes: {
        zh: {
          name: "临时关系",
        },
      },
    });
    const tempRel = respRel.relationship;

    // --- 4.5 List Entities from Event ---
    const now = Math.floor(Date.now() / 1000);
    const yesterday = now - 86400;

    // Test with startNode and time range using event1 (e1) which has location "FANTASY" and tag "产业"
    const { data: list1 } = await api.v1.entityServiceListEntitiesFromEvent({
      startNode: e1?.event?.id,
      startTime: yesterday.toString(),
      endTime: now.toString()
    });
    expect(list1.entities).toBeDefined();

    // Test with other query parameters
    const { data: list2 } = await api.v1.entityServiceListEntitiesFromEvent({
      startNode: e1?.event?.id,
      startTime: yesterday.toString(),
      endTime: now.toString(),
      countryCode: "FANTASY",
      tag: "产业",
      depth: 1
    });
    expect(list2.entities).toBeDefined();

    // --- 5. Delete One of Each ---

    // Delete Temp Relation
    if (tempRel?.key) {
      await api.v1.relationshipServiceDeleteRelationship("event_temp_relation_person", tempRel.key);
    }

    // Delete Event3
    if (e3?.event?.key) {
      await api.v1.entityServiceDeleteEntity("event", e3.event.key);
    }

    // Delete Person3
    if (p3?.person?.key) {
      await api.v1.entityServiceDeleteEntity("person", p3.person.key);
    }

    // Delete Org3
    if (o3?.organization?.key) {
      await api.v1.entityServiceDeleteEntity("organization", o3.organization.key);
    }

    // Delete Web2
    if (w2?.website?.key) {
      await api.v1.entityServiceDeleteEntity("website", w2.website.key);
    }

    // Delete Src3
    if (s3?.source?.key) {
      await api.v1.entityServiceDeleteEntity("source", s3.source.key);
    }
  });
});
