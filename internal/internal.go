package internal

import (
	"strings"
)

//func node2Term(n *Node) *Term {
//id := extractId(n.ID)
//return &Term{
//Id:         id,
//IRI:        n.ID,
//Label:      makeLabel(n, id),
//ChadoId:    makeChadoId(n, id),
//Type:       n.Type,
//Definition: n.Meta.Definition.Val,
//Synonyms:   parseSynonyms(n.Meta.Synonyms),
//Dbxrefs:    parseXrefs(n.Meta.Definition.Xrefs),
//Property:   parseProps(n.Meta.Properties),
//}
//}

//func getCommonTerm(t string) *Term {
//return &Term{
//Id:      t,
//Label:   t,
//ChadoId: t,
//Type:    "PROPERTY",
//}
//}

//func makeChadoId(n *Node, id string) string {
//if n.Type == "CLASS" {
//if strings.Count(id, "_") == 1 {
//return strings.Replace(id, "_", ":", 1)
//}
//}
//return id
//}

//func makeLabel(n *Node, id string) string {
//if len(n.Lbl) == 0 {
//return id
//}
//return n.Lbl
//}

//func parseXrefs(xrefs []string) []*TermDbxref {
//var txrefs []*TermDbxref
//for _, x := range xrefs {
//parts := strings.Split(x, ":")
//txrefs = append(txrefs, &TermDbxref{Database: parts[0], Accession: parts[1]})
//}
//return txrefs
//}

//func parseSynonyms(syn []*Synonym) []*TermSynonym {
//var ts []*TermSynonym
//for _, s := range syn {
//ts = append(
//ts,
//&TermSynonym{
//Name: s.Val,
//Scope: strings.Replace(
//strings.Replace(s.Pred, "has", "", 1),
//"Synonym",
//"",
//1,
//),
//})
//}
//return ts
//}

//func parseProps(p []*Property) *TermProp {
//tp := &TermProp{Deprecated: false}
//for _, v := range p {
//switch {
//case strings.HasSuffix(v.Pred, "#creation_date"):
//t, err := time.Parse(time.RFC3339, v.Val)
//if err == nil {
//tp.CreatedOn = t
//}
//case strings.HasSuffix(v.Pred, "#created_by"):
//tp.CreatedBy = v.Val
//case strings.HasSuffix(v.Pred, "#comment"):
//tp.Comment = v.Val
//case strings.HasSuffix(v.Pred, "#deprecated"):
//tp.Deprecated = true
//case strings.HasSuffix(v.Pred, "#hasOBONamespace"):
//tp.Namespace = v.Val
//case strings.HasSuffix(v.Pred, "IAO_0100001"):
//tp.ReplacedBy = extractId(v.Val)
//case strings.HasSuffix(v.Pred, "#consider"):
//tp.Consider = append(tp.Consider, v.Val)
//default:
//tp.Value = v.Val
//}
//}
//return tp
//}

// ExtractID extracts the last part of an URL primary to create an unique id
// from the IRI values of graph and nodes
func ExtractID(s string) string {
	parts := strings.Split(s, "/")
	l := parts[len(parts)-1]
	if strings.Contains(l, "#") {
		mparts := strings.Split(l, "#")
		return mparts[len(mparts)-1]
	}
	return l
}
