package pan

import (
	"testing"
	"time"
)

type testPost struct {
	ID       int
	Title    string
	Author   int `sql_column:"author_id"`
	Body     string
	Created  time.Time
	Modified *time.Time
}

func (t testPost) GetSQLTableName() string {
	return "test_data"
}

func init() {
	p := testPost{123, "my post", 1, "this is a test post", time.Now(), nil}
	sqlTable[Insert(p)] = queryResult{
		mysql:    "INSERT INTO test_data (test_data.id, test_data.title, test_data.author_id, test_data.body, test_data.created, test_data.modified) VALUES (?,?,?,?,?,?);",
		postgres: "INSERT INTO test_data (test_data.id, test_data.title, test_data.author_id, test_data.body, test_data.created, test_data.modified) VALUES ($1,$2,$3,$4,$5,$6);",
	}
	sqlTable[New("UPDATE "+Table(p)+" SET").Assign(p, "Title", p.Title).Assign(p, "Author", p.Author).Flush(", ").Where().Comparison(p, "ID", "=", p.ID).Flush(" ")] = queryResult{
		mysql:    "UPDATE test_data SET test_data.title = ?, test_data.author_id = ? WHERE test_data.id = ?;",
		postgres: "UPDATE test_data SET test_data.title = $1, test_data.author_id = $2 WHERE test_data.id = $3;",
	}
	sqlTable[New("SELECT "+Columns(p).String()+" FROM "+Table(p)).Where().Expression(Column(p, "Created")+" > (SELECT "+Column(p, "Created")+" FROM "+Table(p)+" WHERE "+Column(p, "ID")+" = ?)", 123).Where().OrderByDesc(Column(p, "Created")).Limit(19).Flush(" ")] = queryResult{
		postgres: "SELECT test_data.id, test_data.title, test_data.author_id, test_data.body, test_data.created, test_data.modified FROM test_data WHERE test_data.created > (SELECT test_data.created FROM test_data WHERE test_data.id = $1) ORDER BY test_data.created DESC LIMIT $2;",
		mysql:    "SELECT test_data.id, test_data.title, test_data.author_id, test_data.body, test_data.created, test_data.modified FROM test_data WHERE test_data.created > (SELECT test_data.created FROM test_data WHERE test_data.id = ?) ORDER BY test_data.created DESC LIMIT ?;",
	}
}

var sqlTable = map[*Query]queryResult{
	New("INSERT INTO "+Table(testPost{})).Expression("("+Placeholders(4)+")", "a", "b", "c", "d").Expression("VALUES").Expression("("+Placeholders(4)+")", 0, 1, 2, 3).Flush(" "): {
		mysql:    "INSERT INTO test_data (?,?,?,?) VALUES (?,?,?,?);",
		postgres: "INSERT INTO test_data ($1,$2,$3,$4) VALUES ($5,$6,$7,$8);",
	},
}

func TestSQLTable(t *testing.T) {
	t.Parallel()
	for query, expectation := range sqlTable {
		t.Logf(query.String())
		mysql, err := query.MySQLString()
		if err != nil {
			t.Errorf("Unexpected error: %+v\n", err)
		}
		postgres, err := query.PostgreSQLString()
		if err != nil {
			t.Errorf("Unexpected error: %+v\n", err)
		}
		if mysql != expectation.mysql {
			t.Errorf("Expected '%s' got '%s'", expectation.mysql, mysql)
		}
		if postgres != expectation.postgres {
			t.Errorf("Expected '%s' got '%s'", expectation.postgres, postgres)
		}
	}
}

func BenchmarkInsertGeneration(b *testing.B) {
	p := testPost{123, "my post", 1, "this is a test post", time.Now(), nil}
	for i := 0; i < b.N; i++ {
		Insert(p)
	}
}

func BenchmarkQueryGeneration(b *testing.B) {
	p := testPost{123, "my post", 1, "this is a test post", time.Now(), nil}
	for i := 0; i < b.N; i++ {
		New("SELECT "+Columns(p).String()+" FROM "+Table(p)).Where().Comparison(p, "ID", "=", p.ID).Expression("OR").In(p, "ID", 123, 456, 789, 101112, 131415).OrderBy(Column(p, "Created")).Limit(10).Flush(" ")
	}
}
