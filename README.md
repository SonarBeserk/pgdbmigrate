Fork of https://github.com/tanel/dbmigrate with Cassandra support stripped out to leave a lightweight PostgreSQL library

Supported databases
-------------------
* PostgreSQL

Install
-------
In your project, place your migrations in a separate folder,
for example, db/migrate.
**Migrations are sorted using their file name and then applied in the sorted order.**
Since sorting is important, name your migrations accordingly. For example,
add a timestamp before migration name. Or use any other ordering scheme you'll like.

Note that migration file names are saved into a table, and the table is used
later on to detect which migrations have already been applied. In other words,
**don't rename your migration files once they've been applied to your DB**.

Use
---

In your app code, import dbmigrate package:
```golang
import (
  "log"
  "path/filepath"

  "github.com/SonarBeserk/pgdbmigrate"
)
```

Then, run the migrations, depending on your database type.

Use with PostgreSQL
-------------------
**Make sure the migrations have an .sql ending.**

After app startup and after a sql.DB instance is initialized in your app, 
run the migrations. Assuming you have a variable called **db** that points to sql.DB
and the migrations are located in **db/migrate**, execute the following code:

```golang
if err := pgdbmigrate.Run(db, filepath.Join("db", "migrate")); err != nil {
  log.Fatal(err)
}
```
