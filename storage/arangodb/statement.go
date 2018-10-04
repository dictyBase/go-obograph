package arangodb

import "fmt"

const (
	tinst = `
		LET fcv = (
			FOR cv IN %s
				FILTER cv.id == "%s"
				RETURN cv._id
		)
		LET existing = (
				FOR cvt IN %s
					FILTER fcv[0] == cvt.graph_id
					RETURN cvt.id
		)
		LET latest = (
			FOR cvt IN %s
				FILTER fcv[0] == cvt.graph_id
				RETURN cvt.id
		)
		FOR diff IN MINUS(latest,existing)
			FOR cvt IN %s
				FILTER diff == cvt.id
				INSERT UNSET(cvt,["_key","_id","_rev"]) IN %s
				COLLECT WITH COUNT INTO c
				RETURN c
	`
	tupdt = `
		LET fcv = (
		    FOR cv in %s
		        FILTER cv.id == "%s"
		        RETURN cv._id
		)
		LET existing = (
	        FOR cvt in %s
	            FILTER fcv[0] == cvt.graph_id
	            RETURN cvt.id
		)
		LET latest = (
			    FOR cvt in %s
		        FILTER fcv[0] == cvt.graph_id
		        RETURN cvt.id
		)
		FOR ins in INTERSECTION(latest,existing)
			    FOR lcvt in %s
			        FOR ecvt in %s
		                FILTER ins == lcvt.id
		                FILTER ins == ecvt.id
		                UPDATE {
		                     _key: ecvt._key,
		                      label: lcvt.label,
		                      metadata: lcvt.metadata
	                   } IN %s
	                   COLLECT WITH COUNT INTO c
	                   RETURN c
		`
)

func termInsert(gname, gcoll, tcoll, temp string) string {
	return fmt.Sprintf(
		tinst,
		gcoll, gname, tcoll, temp, temp, tcoll,
	)
}

func termUpdate(gname, gcoll, tcoll, temp string) string {
	return fmt.Sprintf(
		tupdt,
		gcoll, gname, tcoll, temp, temp, tcoll, tcoll,
	)
}
