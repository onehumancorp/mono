import json

with open("src/test/data/seeders.json", "r") as f:
    data = json.load(f)

data["skillBlueprints"] = [
    {
        "domain": "Digital Marketing",
        "roles": [
            {
                "id": "marketing_director",
                "title": "Marketing Director",
                "context": "Lead the marketing department.",
                "tools": ["mcp://tools/hubspot"]
            },
            {
                "id": "growth_hacker",
                "title": "Growth Hacker",
                "context": "Optimize lead funnels.",
                "reports_to": "marketing_director"
            },
            {
                "id": "content_creator",
                "title": "Content Creator",
                "context": "Write SEO optimized blog posts.",
                "reports_to": "marketing_director"
            }
        ]
    }
]

with open("src/test/data/seeders.json", "w") as f:
    json.dump(data, f, indent=2)
