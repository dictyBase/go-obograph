package arangodb

const (
	getkey = `
		FOR d IN @@db_collection
			FILTER d.id == @db_id
			RETURN d._key
	`
	getid = `
		FOR d IN @@db_collection
			FILTER d.id == @db_id
			RETURN d._id
	`
	getd = `
		FOR d IN @@graph_collection
			FILTER d.id == @graph_id
			RETURN d
	`
	tinst = `
		LET fcv = (
			FOR cv IN @@graph_collection
				FILTER cv.id == @graph_id
				RETURN cv._id
		)
		LET existing = (
				FOR cvt IN @@term_collection
					FILTER fcv[0] == cvt.graph_id
					RETURN cvt.id
		)
		LET latest = (
			FOR cvt IN @@temp_collection
				FILTER fcv[0] == cvt.graph_id
				RETURN cvt.id
		)
		FOR diff IN MINUS(latest,existing)
			FOR cvt IN @@temp_collection
				FILTER diff == cvt.id
				INSERT UNSET(cvt,["_key","_id","_rev"]) IN @@term_collection
				COLLECT WITH COUNT INTO c
				RETURN c
	`
	tupdt = `
		LET fcv = (
			FOR cv IN @@graph_collection
				FILTER cv.id == @graph_id
				RETURN cv._id
		)
		LET existing = (
				FOR cvt IN @@term_collection
					FILTER fcv[0] == cvt.graph_id
					RETURN cvt.id
		)
		LET latest = (
			FOR cvt IN @@temp_collection
				FILTER fcv[0] == cvt.graph_id
				RETURN cvt.id
		)
		FOR ins in INTERSECTION(latest,existing)
			FOR lcvt in @@temp_collection
				FOR ecvt in @@term_collection
					FILTER lcvt.graph_id == fcv[0]
					FILTER ecvt.graph_id == fcv[0]
					FILTER ins == lcvt.id
					FILTER ins == ecvt.id
					UPDATE {
						 _key: ecvt._key,
						  label: lcvt.label,
						  metadata: lcvt.metadata
				   } IN @@term_collection
				   COLLECT WITH COUNT INTO c
				   RETURN c
	`
	tdelt = `
		LET fcv = (
			FOR cv IN @@graph_collection
				FILTER cv.id == @graph_id
				RETURN cv._id
		)
		LET existing = (
				FOR cvt IN @@term_collection
					FILTER fcv[0] == cvt.graph_id
					RETURN cvt.id
		)
		LET latest = (
			FOR cvt IN @@temp_collection
				FILTER fcv[0] == cvt.graph_id
				RETURN cvt.id
		)
		FOR diff in MINUS(existing,latest)
			FOR ecvt in @@term_collection
				FILTER ecvt.graph_id == fcv[0]
				FILTER diff == ecvt.id
				UPDATE {
					 _key: ecvt._key,
					 deprecated: true
				} IN @@term_collection
				COLLECT WITH COUNT INTO c
				RETURN c
	`
	rinst = `
		FOR c IN @@graph_collection
		    FOR cvt IN @@term_collection
		        FILTER c.id == @graph_id
		        FILTER c._id == cvt.graph_id
		        LET nch = MINUS (
		            FOR v IN 1..1 OUTBOUND cvt @@temp_collection
		            OPTIONS { bfs: true, uniqueVertices: 'global' }
		            RETURN v.id,
		            FOR v IN 1..1 OUTBOUND cvt GRAPH @cvterm_graph
		            OPTIONS { bfs: true, uniqueVertices: 'global' }
		            RETURN v.id
		        )
		        FILTER LENGTH(nch) > 0
				FOR n IN nch
					FOR z IN @@temp_collection
		                FOR cvtn IN @@term_collection
		                    FILTER n == cvtn.id
		                    FILTER cvtn._id == z._to
		                    FILTER cvt._id == z._from
		                    INSERT {
		                        _from: z._from,
		                        _to: z._to,
		                        predicate: z.predicate
		                    } IN @@relationship_collection
		                    COLLECT WITH COUNT INTO c
		                    RETURN c
		`
)
