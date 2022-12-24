# IRMS CLI

- Go project to get familiar with the language.
- Allows you to use the IRMS backend through the CLI
- Includes common applications like inserting excel files of records and student lists into the database.

### How to use

- Download the irms.exe file from the directory.
- Set .env in the same folder as the executable. The env variables are as follows
  - WEBSITE - The link to your IRMS backend. e.g. http://localhost:5000/api
- Add the folder to your path ( if needed. You can always run the app from this folder )

### Commands

1. Login - It will ask your password to authenticate.

   ```
   irms login <username / email>
   ```
2. Insert Records

   ```
   irms rec [--semester val] [--subjectcode val] <file path> <sheetname>
   ```
3. Insert Subject

   ```
   irms sub <subject code> <subject name>
   ```
4. Insert Students

   ```
   irms register --degree <val> <file path> <sheet name>
   ```
