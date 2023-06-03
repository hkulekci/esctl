package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fehmicansaglam/esctl/constants"
	"github.com/fehmicansaglam/esctl/es"
	"github.com/fehmicansaglam/esctl/output"
	"github.com/spf13/cobra"
)

var (
	flagIndex        string
	flagNode         string
	flagShard        int
	flagPrimary      bool
	flagReplica      bool
	flagStarted      bool
	flagRelocating   bool
	flagInitializing bool
	flagUnassigned   bool
	flagActions      []string
	flagSortBy       []string
)

var getCmd = &cobra.Command{
	Use:   "get ENTITY",
	Short: "Get Elasticsearch entities",
	Long: trim(`
The 'get' command allows you to retrieve information about Elasticsearch entities.

Available Entities:
  - nodes: List all nodes in the Elasticsearch cluster.
  - indices: List all indices in the Elasticsearch cluster.
  - shards: List detailed information about shards, including their sizes and placement.
  - aliases: List all aliases in the Elasticsearch cluster.
  - tasks: List all tasks in the Elasticsearch cluster.`),
	Example: trimAndIndent(`
#Retrieve a list of all nodes in the Elasticsearch cluster.
esctl get nodes

#Retrieve a list of all indices in the Elasticsearch cluster.
esctl get indices

#Retrieve detailed information about shards in the Elasticsearch cluster.
esctl get shards

#Retrieve shard information for an index.
esctl get shards --index my_index

#Retrieve shard information filtered by state.
esctl get shards --started --relocating

#Retrieve all aliases.
esctl get aliases

#Retrieve tasks filtered by actions using wildcard patterns.
esctl get tasks --actions 'index*' --actions '*search*'

#Retrieve all tasks.
esctl get tasks`),
	Args:       cobra.ExactArgs(1),
	ValidArgs:  []string{"node", "index", "shard", "alias", "task"},
	ArgAliases: []string{"nodes", "indices", "shards", "aliases", "tasks"},
	Run: func(cmd *cobra.Command, args []string) {
		entity := args[0]

		switch entity {
		case constants.EntityNode, constants.EntityNodes:
			handleNodeLogic()
		case constants.EntityIndex, constants.EntityIndices:
			handleIndexLogic()
		case constants.EntityShard, constants.EntityShards:
			handleShardLogic()
		case constants.EntityAlias, constants.EntityAliases:
			handleAliasLogic()
		case constants.EntityTask, constants.EntityTasks:
			handleTaskLogic()
		default:
			fmt.Printf("Unknown entity: %s\n", entity)
			os.Exit(1)
		}
	},
}

func handleNodeLogic() {
	nodes, err := es.GetNodes()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to retrieve nodes:", err)
		os.Exit(1)
	}

	columnDefs := []output.ColumnDef{
		{Header: "NAME", Type: output.Text},
		{Header: "IP", Type: output.Text},
		{Header: "NODE-ROLE", Type: output.Text},
		{Header: "MASTER", Type: output.Text},
		{Header: "HEAP-MAX", Type: output.DataSize},
		{Header: "HEAP-CURRENT", Type: output.DataSize},
		{Header: "HEAP-PERCENT", Type: output.Percent},
		{Header: "CPU", Type: output.Percent},
		{Header: "LOAD-1M", Type: output.Number},
		{Header: "DISK-TOTAL", Type: output.DataSize},
		{Header: "DISK-USED", Type: output.DataSize},
		{Header: "DISK-AVAILABLE", Type: output.DataSize},
	}
	data := [][]string{}

	for _, node := range nodes {
		row := []string{
			node.Name, node.IP, node.NodeRole, node.Master, node.HeapMax, node.HeapCurrent,
			node.HeapPercent + "%", node.CPU + "%", node.Load1m,
			node.DiskTotal, node.DiskUsed, node.DiskAvail,
		}
		data = append(data, row)
	}

	if len(flagSortBy) > 0 {
		output.PrintTable(columnDefs, data, flagSortBy...)
	} else {
		output.PrintTable(columnDefs, data, "NAME")
	}
}

func handleIndexLogic() {
	indices, err := es.GetIndices(flagIndex)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to retrieve indices:", err)
		os.Exit(1)
	}

	columnDefs := []output.ColumnDef{
		{Header: "INDEX", Type: output.Text},
		{Header: "UUID", Type: output.Text},
		{Header: "HEALTH", Type: output.Text},
		{Header: "STATUS", Type: output.Text},
		{Header: "SHARDS", Type: output.Number},
		{Header: "REPLICAS", Type: output.Number},
		{Header: "DOCS-COUNT", Type: output.Number},
		{Header: "DOCS-DELETED", Type: output.Number},
		{Header: "CREATION-DATE", Type: output.Date},
		{Header: "STORE-SIZE", Type: output.DataSize},
		{Header: "PRI-STORE-SIZE", Type: output.DataSize},
	}
	data := [][]string{}

	for _, index := range indices {
		row := []string{
			index.Index, index.UUID, index.Health, index.Status, index.Pri, index.Rep,
			index.DocsCount, index.DocsDeleted, index.CreationDate, index.StoreSize, index.PriStoreSize,
		}
		data = append(data, row)
	}

	if len(flagSortBy) > 0 {
		output.PrintTable(columnDefs, data, flagSortBy...)
	} else {
		output.PrintTable(columnDefs, data, "INDEX")
	}
}

func includeShardByState(shard es.Shard) bool {
	switch {
	case flagStarted && shard.State == constants.ShardStateStarted:
		return true
	case flagRelocating && shard.State == constants.ShardStateRelocating:
		return true
	case flagInitializing && shard.State == constants.ShardStateInitializing:
		return true
	case flagUnassigned && shard.State == constants.ShardStateUnassigned:
		return true
	case !flagStarted && !flagRelocating && !flagInitializing && !flagUnassigned:
		return true
	}
	return false
}

func includeShardByNumber(shard es.Shard) bool {
	shardNumber, err := strconv.Atoi(shard.Shard)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse shard number:", err)
		os.Exit(1)
	}
	return flagShard == -1 || flagShard == shardNumber
}

func includeShardByPriRep(shard es.Shard) bool {
	return (flagPrimary && shard.PriRep == constants.ShardPrimary) ||
		(flagReplica && shard.PriRep == constants.ShardReplica) ||
		(!flagPrimary && !flagReplica)
}

func includeShardByNode(shard es.Shard) bool {
	if flagNode == "" {
		return true
	}

	return shard.Node == flagNode
}

func handleShardLogic() {
	shards, err := es.GetShards(flagIndex)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to retrieve shards:", err)
		os.Exit(1)
	}

	columnDefs := []output.ColumnDef{
		{Header: "INDEX", Type: output.Text},
		{Header: "SHARD", Type: output.Number},
		{Header: "PRI-REP", Type: output.Text},
		{Header: "STATE", Type: output.Text},
		{Header: "DOCS", Type: output.Number},
		{Header: "STORE", Type: output.DataSize},
		{Header: "IP", Type: output.Text},
		{Header: "NODE", Type: output.Text},
		{Header: "NODE-ID", Type: output.Text},
		{Header: "UNASSIGNED-REASON", Type: output.Text},
		{Header: "UNASSIGNED-AT", Type: output.Date},
		{Header: "SEGMENTS-COUNT", Type: output.Number},
	}
	data := [][]string{}

	for _, shard := range shards {
		if includeShardByState(shard) && includeShardByNumber(shard) && includeShardByPriRep(shard) && includeShardByNode(shard) {
			row := []string{
				shard.Index, shard.Shard, humanizePriRep(shard.PriRep), shard.State, shard.Docs, shard.Store, shard.IP, shard.Node, shard.ID, shard.UnassignedReason, shard.UnassignedAt, shard.SegmentsCount,
			}
			data = append(data, row)
		}
	}

	if len(flagSortBy) > 0 {
		output.PrintTable(columnDefs, data, flagSortBy...)
	} else {
		output.PrintTable(columnDefs, data, "INDEX", "SHARD", "PRI-REP")
	}
}

func humanizePriRep(priRep string) string {
	switch priRep {
	case constants.ShardPrimary:
		return "primary"
	case constants.ShardReplica:
		return "replica"
	default:
		return priRep
	}
}

func handleAliasLogic() {
	aliases, err := es.GetAliases(flagIndex)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to retrieve aliases:", err)
		os.Exit(1)
	}

	columnDefs := []output.ColumnDef{
		{Header: "ALIAS", Type: output.Text},
		{Header: "INDEX", Type: output.Text},
	}
	data := [][]string{}

	for alias, index := range aliases {
		row := []string{alias, index}
		data = append(data, row)
	}

	output.PrintTable(columnDefs, data, flagSortBy...)
}

func handleTaskLogic() {
	tasksResponse, err := es.GetTasks(flagActions)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to retrieve tasks:", err)
		os.Exit(1)
	}

	columnDefs := []output.ColumnDef{
		{Header: "NODE", Type: output.Text},
		{Header: "ID", Type: output.Number},
		{Header: "ACTION", Type: output.Text},
	}
	data := [][]string{}

	for _, node := range tasksResponse.Nodes {
		for _, task := range node.Tasks {
			row := []string{task.Node, fmt.Sprintf("%d", task.ID), task.Action}
			data = append(data, row)
		}
	}

	if len(flagSortBy) > 0 {
		output.PrintTable(columnDefs, data, flagSortBy...)
	} else {
		output.PrintTable(columnDefs, data, "NODE", "ID")
	}
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVar(&flagIndex, "index", "", "Name of the index")
	getCmd.Flags().StringVar(&flagNode, "node", "", "Filter shards by node")
	getCmd.Flags().IntVar(&flagShard, "shard", -1, "Filter shards by shard number")
	getCmd.Flags().BoolVar(&flagPrimary, "primary", false, "Filter primary shards")
	getCmd.Flags().BoolVar(&flagReplica, "replica", false, "Filter replica shards")
	getCmd.Flags().BoolVar(&flagStarted, "started", false, "Filter shards in STARTED state")
	getCmd.Flags().BoolVar(&flagRelocating, "relocating", false, "Filter shards in RELOCATING state")
	getCmd.Flags().BoolVar(&flagInitializing, "initializing", false, "Filter shards in INITIALIZING state")
	getCmd.Flags().BoolVar(&flagUnassigned, "unassigned", false, "Filter shards in UNASSIGNED state")
	getCmd.Flags().StringSliceVar(&flagActions, "actions", []string{}, "Filter tasks by actions")
	getCmd.Flags().StringSliceVar(&flagSortBy, "sort-by", []string{}, "Columns to sort by (comma-separated)")
}
