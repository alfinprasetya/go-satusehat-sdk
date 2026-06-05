# SATUSEHAT FHIR Postman Guidelines

Use this directory to store Postman assets copied or adapted from the official SATUSEHAT FHIR documentation. These files are intended as implementation guidelines for SDK development, request mapping, and manual API exploration.

## Directory layout

```text
docs/postman
├── collections
│   ├── 00. FHIR Resource - Contoh Penggunaan.postman_collection.json
│   └── README.md
└── README.md
```

## Collection source

The official SATUSEHAT FHIR R4 Postman collection is saved at `collections/00. FHIR Resource - Contoh Penggunaan.postman_collection.json`.

Official source: [00. FHIR Resource - Contoh Penggunaan](https://www.postman.com/satusehat/satusehat-public/collection/u2k8uiz/00-fhir-resource-contoh-penggunaan)

Use the checked-in collection as the source guideline for SDK endpoint paths, query parameters, request bodies, and expected responses. To refresh it, export the latest collection from Postman and replace the JSON file in `collections/`.

## Notes

- Prefer official SATUSEHAT examples over hand-written examples when both exist.
- Keep collection exports readable where possible by avoiding unnecessary Postman metadata churn.
- Do not commit generated access tokens, refresh tokens, client secrets, or production identifiers.
