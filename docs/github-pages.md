---
title: GitHub Pages
nav_order: 11
---

## Setup GitHub Pages untuk folder `docs/`

Dokumentasi ini dibuat agar bisa langsung dipublish via GitHub Pages menggunakan **Jekyll theme bawaan** (tanpa install tools tambahan).

### 1. Pastikan file berikut ada

- `docs/index.md`
- `docs/_config.yml`

### 2. Aktifkan GitHub Pages

Di GitHub repo:

- Buka **Settings → Pages**
- **Source**: `Deploy from a branch`
- **Branch**: pilih branch utama (mis. `main`)
- **Folder**: pilih `/docs`
- Save

Tunggu beberapa menit sampai build selesai.

### 3. URL

GitHub akan menampilkan URL Pages di halaman yang sama (Settings → Pages).

### 4. Tips

- Untuk asset/gambar, simpan di `docs/assets/` lalu refer dengan path relatif.
- Jika kamu butuh custom domain, set di halaman Pages lalu tambahkan file `docs/CNAME`.

