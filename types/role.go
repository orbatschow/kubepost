package types

type PostgresRole struct {
	Rolname        string
	Rolsuper       bool
	Rolinherit     bool
	Rolcreaterole  bool
	Rolcreatedb    bool
	Rolcanlogin    bool
	Rolreplication bool
	Rolconnlimit   int
	Rolpassword    string
	Rolvaliduntil  string
	Rolbypassrls   bool
	Rolconfig      string
	Oid            int
}
