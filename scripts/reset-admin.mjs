import pg from 'pg'
import bcrypt from 'bcryptjs'

const email = (process.env.ADMIN_EMAIL || 'admin@cybermate.bot').trim().toLowerCase()
const password = process.env.ADMIN_PASSWORD || 'ВашПароль123!'

function dbConfig() {
  if (process.env.DATABASE_URL) {
    return {
      connectionString: process.env.DATABASE_URL,
      ssl: { rejectUnauthorized: false },
    }
  }

  return {
    host: process.env.PG_HOST || '24879e6f20791375a1bc3a29.twc1.net',
    port: Number(process.env.PG_PORT || 5432),
    user: process.env.PG_USER || 'gen_user',
    password: process.env.PG_PASSWORD || process.env.PG_PASS || 'ebmIWt4CCv}Vsg',
    database: process.env.PG_DBNAME || 'default_db',
    ssl: { rejectUnauthorized: false },
  }
}

const client = new pg.Client(dbConfig())

try {
  await client.connect()

  await client.query(`
    CREATE TABLE IF NOT EXISTS admins (
      id BIGSERIAL PRIMARY KEY,
      email TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'admin',
      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    )
  `)

  const hash = await bcrypt.hash(password, 10)
  const result = await client.query(
    `INSERT INTO admins(email, password_hash, role)
     VALUES($1, $2, 'admin')
     ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
     RETURNING id, email`,
    [email, hash],
  )

  const row = result.rows[0]
  console.log(`admin upserted: ${row.email} (id=${row.id})`)
} catch (err) {
  console.error('reset failed:', err.message)
  process.exit(1)
} finally {
  await client.end()
}
