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
        owner: "admin",
        read: [],
        write: [],
        title: "Rainbow Unicorn Supply Chain Magic Conference",
        type: "Magic Conference",
        location: {
          latitude: 36.8835,
          longitude: -123.43,
          countryCode: "FANTASY",
          administrativeArea: "Rainbow Kingdom",
          subAdministrativeArea: "Magic Forest County",
          locality: "Candy Castle",
          subLocality: "Rainbow Avenue",
          address: "123 Rainbow Unicorn Street",
          postalCode: 99999,
        },
        description: "Hosted by Rainbow Unicorn, this conference discusses how to restructure the supply chain with magic powder, transport goods using the Rainbow Bridge, and adjust tariff policies with magic wands. Attendees include talking animals, flying cars, and dancing trees, exploring fairytale solutions to real-world problems.",
        tags: ["Industry", "Supply Chain", "Tariff"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
        attributes: {
          zh: {
            title: "彩虹独角兽供应链魔法大会",
            type: "魔法大会",
            description: "本次大会由彩虹独角兽主办，讨论如何用魔法粉末重组供应链，使用彩虹桥运输货物，以及如何用魔法棒调整关税政策。与会者包括会说话的动物、会飞的汽车和会跳舞的树木，共同探讨用童话方式解决现实问题。",
            tags: ["产业", "供应链", "关税"],
          },
        },
      }
    };
    const { data: respE1 } = await api.v1.entityServiceCreateEntity("event", event1);
    const e1 = respE1.entity;
    expect(e1?.event?.id).toBeDefined();

    const event2: V1Entity = {
      event: {
        owner: "admin",
        read: [],
        write: [],
        title: "Interstellar Trade Expo",
        type: "Expo",
        location: {
          latitude: 40.7128,
          longitude: -74.0060,
          countryCode: "SPACE",
          administrativeArea: "Milky Way",
          subAdministrativeArea: "Solar System",
          locality: "Mars Colony",
          subLocality: "Red Plains",
          address: "42 Mars One Avenue",
          postalCode: 0,
        },
        description: "Merchants and explorers from all over the universe gather together to showcase the latest light-speed spaceships, quantum communication devices, and anti-gravity boots. Key discussions include how to establish an intergalactic free trade zone and how to deal with logistics delays caused by black holes.",
        tags: ["Technology", "Trade", "Interstellar"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
        attributes: {
          zh: {
            title: "星际穿越贸易博览会",
            type: "博览会",
            description: "来自全宇宙的商人和探险家齐聚一堂，展示最新的光速飞船、量子通讯设备和反重力靴子。重点讨论如何建立跨星系自由贸易区，以及如何应对黑洞造成的物流延误。",
            tags: ["科技", "贸易", "星际"],
          },
        },
      }
    };
    const { data: respE2 } = await api.v1.entityServiceCreateEntity("event", event2);
    const e2 = respE2.entity;

    const event3: V1Entity = {
      event: {
        owner: "admin",
        read: [],
        write: [],
        title: "Deep Sea Rare Treasures Auction",
        type: "Auction",
        location: {
          latitude: -25.2744,
          longitude: 133.7751,
          countryCode: "OCEAN",
          administrativeArea: "Pacific Ocean",
          subAdministrativeArea: "Mariana Trench",
          locality: "Atlantis Ruins",
          subLocality: "Coral Plaza",
          address: "888 Deep Sea Avenue",
          postalCode: 88888,
        },
        description: "A mysterious deep-sea auction featuring items such as mermaid tears, ink paintings by giant octopuses, and ancient gold coins from shipwrecks. Proceeds from this auction will be used to protect the marine ecosystem and prevent plastic pollution.",
        tags: ["Auction", "Treasure", "Environmental Protection"],
        happenedAt: Math.floor(Date.now() / 1000).toString(),
        attributes: {
          zh: {
            title: "深海奇珍异宝拍卖会",
            type: "拍卖会",
            description: "神秘的深海拍卖会，拍品包括美人鱼的眼泪、巨型章鱼的墨汁画和沉船中的古老金币。本次拍卖会所得款项将用于保护海洋生态环境，防止塑料垃圾污染海洋。",
            tags: ["拍卖", "珍宝", "环保"],
          },
        },
      }
    };
    const { data: respE3 } = await api.v1.entityServiceCreateEntity("event", event3);
    const e3 = respE3.entity;

    // Persons
    const person1: V1Entity = {
      person: {
        owner: "admin",
        read: [],
        write: [],
        name: "Gandalf the White",
        role: "Wizard",
        nationality: "Maia",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["Wizard", "Leader", "Magic"],
        aliases: ["WG"],
        attributes: {
          zh: {
            name: "白袍甘道夫",
            role: "巫师",
            nationality: "迈雅",
            tags: ["巫师", "领袖", "魔法"],
          },
        },
      }
    };
    const { data: respP1 } = await api.v1.entityServiceCreateEntity("person", person1);
    const p1 = respP1.entity;

    const person2: V1Entity = {
      person: {
        owner: "admin",
        read: [],
        write: [],
        name: "Elon Musk",
        role: "Entrepreneur",
        nationality: "Martian",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["Entrepreneur", "Technology", "Space"],
        aliases: ["EM"],
        attributes: {
          zh: {
            name: "埃隆·马斯",
            role: "企业家",
            nationality: "火星人",
            tags: ["企业家", "科技", "太空"],
          },
        },
      }
    };
    const { data: respP2 } = await api.v1.entityServiceCreateEntity("person", person2);
    const p2 = respP2.entity;

    const person3: V1Entity = {
      person: {
        owner: "admin",
        read: [],
        write: [],
        name: "Captain Nemo",
        role: "Captain",
        nationality: "Indian",
        birthDate: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["Captain", "Explorer", "Ocean"],
        aliases: ["Nemo"],
        attributes: {
          zh: {
            name: "尼莫船长",
            role: "船长",
            nationality: "印度",
            tags: ["船长", "探险家", "海洋"],
          },
        },
      }
    };
    const { data: respP3 } = await api.v1.entityServiceCreateEntity("person", person3);
    const p3 = respP3.entity;

    // Organizations
    const org1: V1Entity = {
      organization: {
        owner: "admin",
        read: [],
        write: [],
        name: "Unicorn Supply Chain Co.",
        type: "For-Profit Company",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["Logistics", "Magic", "Supply Chain"],
        attributes: {
          zh: {
            name: "独角兽供应链公司",
            type: "盈利性公司",
            tags: ["物流", "魔法", "供应链"],
          },
        },
      }
    };
    const { data: respO1 } = await api.v1.entityServiceCreateEntity("organization", org1);
    const o1 = respO1.entity;

    const org2: V1Entity = {
      organization: {
        owner: "admin",
        read: [],
        write: [],
        name: "Interstellar Explorers",
        type: "For-Profit Company",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["Spaceflight", "Technology", "Exploration"],
        attributes: {
          zh: {
            name: "星际探索者",
            type: "盈利性公司",
            tags: ["航天", "科技", "探索"],
          },
        },
      }
    };
    const { data: respO2 } = await api.v1.entityServiceCreateEntity("organization", org2);
    const o2 = respO2.entity;

    const org3: V1Entity = {
      organization: {
        owner: "admin",
        read: [],
        write: [],
        name: "Deep Blue Conservation Society",
        type: "Non-Profit Organization",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["Environmental Protection", "Ocean", "Charity"],
        attributes: {
          zh: {
            name: "深蓝保护协会",
            type: "非营利组织",
            tags: ["环保", "海洋", "公益"],
          },
        },
      }
    };
    const { data: respO3 } = await api.v1.entityServiceCreateEntity("organization", org3);
    const o3 = respO3.entity;

    // Websites
    const web1: V1Entity = {
      website: {
        owner: "admin",
        read: [],
        write: [],
        url: "https://www.magic-supply-chain.fantasy",
        title: "Unicorn Supply Chain Official Website",
        description: "The official website of Unicorn Supply Chain Company, covering everything from supply chain logistics to magic business.",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["Magic", "Business"],
        attributes: {
          zh: {
            title: "独角兽供应链官方网",
            description: "独角兽供应链公司的官方网站，涵盖从供应链物流到魔法业务的所有内容。",
            tags: ["魔法", "商业"],
          },
        },
      }
    };
    const { data: respW1 } = await api.v1.entityServiceCreateEntity("website", web1);
    const w1 = respW1.entity;

    const web2: V1Entity = {
      website: {
        owner: "admin",
        read: [],
        write: [],
        url: "https://www.mars-colonies.space",
        title: "Mars Colony Express",
        description: "Latest news about Mars colony construction, life, and trade.",
        foundedAt: Math.floor(Date.now() / 1000).toString(),
        discoveredAt: Math.floor(Date.now() / 1000).toString(),
        lastVisited: Math.floor(Date.now() / 1000).toString(),
        tags: ["Space", "News"],
        attributes: {
          zh: {
            title: "火星殖民地快讯",
            description: "关于火星殖民地建设、生活和贸易的最新消息。",
            tags: ["太空", "新闻"],
          },
        },
      }
    };
    const { data: respW2 } = await api.v1.entityServiceCreateEntity("website", web2);
    const w2 = respW2.entity;

    // Sources
    const src1: V1Entity = {
      source: {
        owner: "admin",
        read: [],
        write: [],
        name: "The Daily Prophet",
        type: "News",
        url: "https://www.thedailyprophet.com/article-one",
        reliability: 20,
        title: "Headline News: The Rise of Magic Supply Chain",
        description: "The Daily Prophet released the latest news on the magic supply chain, which will change the operating model of the entire magic industry.",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["News", "Magic"],
        attributes: {
          zh: {
            name: "预言家日报",
            type: "新闻",
            title: "头条新闻：魔法供应链的崛起",
            description: "预言家日报发布了关于魔法供应链的最新消息，这将改变整个魔法行业的运营模式。",
            tags: ["新闻", "魔法"],
          },
        },
      }
    };
    const { data: respS1 } = await api.v1.entityServiceCreateEntity("source", src1);
    const s1 = respS1.entity;

    const src2: V1Entity = {
      source: {
        owner: "admin",
        read: [],
        write: [],
        name: "Galactic Gazette",
        type: "News",
        url: "https://www.galactic-gazette.space/daily-news",
        reliability: 80,
        title: "Galactic Trade Agreement Signed",
        description: "The Galactic Gazette released the latest news on the signing of the Galactic Trade Agreement, which will change the trade model of the entire universe.",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["News", "Space"],
        attributes: {
          zh: {
            name: "银河公报",
            type: "新闻",
            title: "银河系贸易协定签署",
            description: "银河公报发布了关于银河系贸易协定签署的最新消息，这将改变整个宇宙的贸易模式。",
            tags: ["新闻", "太空"],
          },
        },
      }
    };
    const { data: respS2 } = await api.v1.entityServiceCreateEntity("source", src2);
    const s2 = respS2.entity;

    const src3: V1Entity = {
      source: {
        owner: "admin",
        read: [],
        write: [],
        name: "Deep Sea Echo",
        type: "News",
        url: "https://www.deep-sea-echo.ocean/reports",
        reliability: 90,
        title: "New Discovery at Atlantis Ruins",
        description: "The Deep Sea Echo team has made a new discovery at the Atlantis Ruins, a mysterious island containing ancient magical powers.",
        createdAt: Math.floor(Date.now() / 1000).toString(),
        updatedAt: Math.floor(Date.now() / 1000).toString(),
        tags: ["News", "Ocean"],
        attributes: {
          zh: {
            name: "深海回声",
            type: "新闻",
            title: "亚特兰蒂斯遗址的新发现",
            description: "深海回声团队发现了亚特兰蒂斯遗址的新发现，这是一个神秘的岛屿，包含了古代的魔法力量。",
            tags: ["新闻", "海洋"],
          },
        },
      }
    };
    const { data: respS3 } = await api.v1.entityServiceCreateEntity("source", src3);
    const s3 = respS3.entity;

    // --- 4. Connect Entities ---
    await api.v1.relationshipServiceCreateRelationship({
      from: e1?.event?.id,
      to: p1?.person?.id,
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
      owner: "admin",
      read: [],
      write: [],
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
