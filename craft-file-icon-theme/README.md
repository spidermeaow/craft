# Craft File Icon 0.2.0

Version 0.2.0 supersedes the historical theme described below. It registers ONLY a default language icon for `.craft`, using the supplied cyan logo. No iconThemes, generic files, folders or config/lock mappings are contributed. Historical assets/generator are excluded from the VSIX.

Upgrade: Install the 0.2.0 VSIX, then select your previous theme using Preferences: File Icon Theme (Seti, Material Icon Theme, etc.) and reload the window. Do not select the retired Craft Forge theme. No settings are changed automatically.

Craft Language Support 0.2.2 contains the same icon; the standalone extension is optional. Themes that override Craft or disable language icons may hide the logo. Other file icons remain controlled by the selected theme.

User check: compare .craft, .js, .py, craft.toml, craft.lock and folders under light/dark themes. Only Craft language files receive this icon. UI tests have not been run.

## Historical 0.1.0 design (retired)

ชุด File Icon Theme สำหรับโปรเจกต์ Craft 0.1.x และเครื่องมือที่สร้างต่อจากภาษา Craft

## แนวทางภาพ

- `.craft` ใช้สัญลักษณ์ประกายไฟของ Craft สื่อถึง source code
- `craft.toml` ใช้เฟือง/โครงสร้าง สื่อถึง project configuration
- `craft.lock` ใช้แม่กุญแจ สื่อถึง dependency lock
- `test` ใช้เครื่องหมายถูก สื่อถึงการตรวจสอบความถูกต้อง
- `migration` ใช้ลูกศรหมุน สื่อถึงการเปลี่ยน schema แบบเป็นลำดับ
- `build` ใช้กล่องผลลัพธ์ สื่อถึง executable artifact

## โครงสร้าง

```text
craft-file-icon-theme/
├── theme.json
├── README.md
└── icons/
    ├── craft.svg
    ├── config.svg
    ├── lock.svg
    ├── test.svg
    ├── migration.svg
    └── build.svg
```

## การนำไปใช้

ไฟล์ SVG เป็น vector และไม่มี dependency ภายนอก เหมาะสำหรับนำไป map กับ VS Code, editor ของ Craft หรือ file explorer ของ IDE อื่น

`theme.json` เป็น manifest กลางสำหรับให้ Codex อ่าน mapping และนำไปสร้าง adapter เฉพาะแพลตฟอร์มต่อไป


## VS Code extension

ใช้ SVG เดิมทั้ง 6 ไฟล์ เพิ่ม generic file/folder ใน `adapter-icons/`
`theme.json` เป็น source of truth; `generated/` เป็น adapter ที่สร้างจาก manifest
Extension ID สำหรับ local VSIX คือ `craft-local.craft-forge-file-icons` (ไม่ใช่ Marketplace publisher ที่จดทะเบียน)
รองรับ VS Code 1.85 ขึ้นไป ไม่ต้องมี Craft CLI หรือ Go และไม่มี activation code

### Build

```powershell
cd craft-file-icon-theme
npm run generate
npx --yes @vscode/vsce@4.0.0 package --allow-missing-repository --out ../dist/craft-forge-file-icons-0.1.0.vsix
```

### ติดตั้งและตรวจรับด้วยตนเอง

1. VS Code: Extensions → `...` → Install from VSIX → เลือกไฟล์ใน `dist/`
2. Command Palette → Preferences: File Icon Theme → Craft Forge File Icons
3. เปิด `fixtures/` ตรวจตามตาราง ทั้ง Light/Dark/High Contrast และขนาด UI ต่าง ๆ
4. ตรวจ generic icons ของไฟล์/โฟลเดอร์ที่ไม่ใช่ Craft
5. เปลี่ยนกลับด้วย File Icon Theme หรือถอน Craft Forge File Icons จาก Extensions

ทางเลือก: `code --install-extension <path-to-vsix>` และ `code --uninstall-extension craft-local.craft-forge-file-icons`
การเลือก theme ไม่ได้รวมไอคอนจาก theme เดิม ไม่เปลี่ยน Windows Explorer หรือ default editor

| ตัวอย่าง | Icon ที่คาดหวัง |
| --- | --- |
| src/main.craft | Craft |
| craft.toml | Config |
| craft.lock | Lock |
| src/account_test.craft | Craft (suffix wildcard fallback) |
| tests/account.craft, tests/account_test.craft | Test |
| tests/unit/account.craft | Craft (parent คือ unit) |
| nested/tests/account.craft | Test (parent ชื่อ tests ที่ระดับใดก็ได้) |
| migrations/001.craft | Migration |
| build/, dist/ ทั้งพับและเปิด | Build |
| notes.txt, other/ | Generic file/folder |

`*_test.craft` ใช้ native VS Code mapping ไม่ได้
`tests/*.craft` / `migrations/*.craft` เป็น parent โดยตรง ไม่ recursive และไม่จำกัด project root
ดู `generated/association-report.json` สำหรับกฎที่ข้ามและ fallback
ไอคอน lock/migration ไม่ได้หมายความว่ามี dependency manager หรือ database migration แล้ว

คำสั่งตรวจ generated adapter สำหรับผู้ใช้: `npm run check`
Agent ยังไม่ได้ตรวจการแสดงผล/ติดตั้งใน VS Code จริง
JetBrains, Craft editor และ CLI adapter อยู่ใน backlog

License: MIT ตาม manifest และการยืนยันของผู้ใช้; ไม่ระบุชื่อเจ้าของที่ยังไม่ได้รับข้อมูล

อ้างอิง adapter: [VS Code File Icon Theme API](https://code.visualstudio.com/api/extension-guides/file-icon-theme) และ [parent-folder mapping ใน VS Code 1.85](https://github.com/microsoft/vscode/blob/1.85.0/src/vs/workbench/services/themes/browser/fileIconThemeData.ts)
