package throughcache

func QueryerRelations(querys []Queryer) ([]string, []string, map[string]Queryer) {
	keys := make([]string, 0, len(querys))
	ids := make([]string, 0, len(querys))
	id2Query := make(map[string]Queryer)
	for _, q := range querys {
		keys = append(keys, q.MakeKey())
		ids = append(ids, q.ID())
		id2Query[q.ID()] = q
	}
	return keys, ids, id2Query
}

func Keys(querys []Queryer) []string {
	keys, _, _ := QueryerRelations(querys)
	return keys
}

func IDs(querys []Queryer) []string {
	_, ids, _ := QueryerRelations(querys)
	return ids
}

func Id2Queryer(querys []Queryer) map[string]Queryer {
	_, _, Id2Queryer := QueryerRelations(querys)
	return Id2Queryer
}

func Id2Modeler(modelers []Modeler) map[string]Modeler {
	id2Modeler := make(map[string]Modeler)
	for _, modeler := range modelers {
		id2Modeler[modeler.ID()] = modeler
	}
	return id2Modeler
}
