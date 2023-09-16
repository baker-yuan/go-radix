package radix

import (
	"sort"
	"strings"
)

// WalkFn is used when walking the tree. Takes a
// key and value, returning if iteration should
// be terminated.
type WalkFn func(s string, v interface{}) bool

// leafNode is used to represent a value
//
// 表示一个叶子节点。它有两个字段：key（键）和val（值）。
type leafNode struct {
	key string
	val interface{}
}

// edge is used to represent an edge node
//
// 表示一个边节点。它有两个字段：label（标签）和node（节点）。
type edge struct {
	// 在这个`edge`结构体中，有两个字段：`label`和`node`。
	// 1. `label`：这是一个字节类型的字段，代表了边的标签。在许多树形数据结构中，特别是在前缀树（Trie）或基数树（Radix Tree）中，边通常用于表示从一个节点到另一个节点的路径。这个路径可以是一个字符，一个字符串，或者其他类型的标签。在这个特定的实现中，`label`可能代表了从父节点到这个节点的路径上的一个字符。
	// 2. `node`：这是一个指向`node`类型的指针，代表了这个边的终点节点。在树形数据结构中，一个边（Edge）通常用于连接两个节点（Node），一个是起点节点（通常是父节点），一个是终点节点（通常是子节点）。这个`node`字段的主要作用是存储这个边所连接的节点的信息。通过这个字段，你可以从一个节点沿着这个边到达另一个节点。
	// 总的来说，`label`字段用于标识和区分不同的边，而`node`字段用于存储和表示这个边所连接的节点。

	label byte
	node  *node
}

// 表示一个节点。它有三个字段：leaf（叶子节点）、prefix（前缀）和edges（边）。
// leaf字段用于存储可能的叶子节点，prefix字段用于存储我们忽略的公共前缀，edges字段用于存储边，这些边应该按顺序存储以便于迭代。
type node struct {
	// leaf is used to store possible leaf
	// 指向leafNode类型的指针，用于存储可能的叶子节点
	leaf *leafNode

	// prefix is the common prefix we ignore
	// 忽略的公共前缀。
	prefix string

	// Edges should be stored in-order for iteration.
	// We avoid a fully materialized slice to save memory,
	// since in most cases we expect to be sparse
	// 存储边。边应该按照顺序存储以便于迭代。为了节省内存，我们避免了完全实体化的切片，因为在大多数情况下，我们预期它们会是稀疏的。
	edges edges
}

// 检查当前节点是否是叶子节点。
func (n *node) isLeaf() bool {
	return n.leaf != nil
}

// 在当前节点添加一个边
// 这个方法接受一个edge类型的参数e。它没有返回值。
func (n *node) addEdge(e edge) {
	// 它获取当前节点边的数量num
	num := len(n.edges)
	// 使用sort.Search函数在边中查找给定的标签。
	// sort.Search函数接受两个参数：一个是要搜索的元素的数量，另一个是一个函数，这个函数接受一个索引并返回一个布尔值，表示是否找到了要搜索的元素。
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= e.label
	})
	// 在边的切片中添加一个新的边，并将从找到的索引开始的所有边向后移动一位。
	n.edges = append(n.edges, edge{})
	copy(n.edges[idx+1:], n.edges[idx:])
	// 在找到的索引处插入新的边。
	n.edges[idx] = e
}

// 更新当前节点的一个边
// 方法接受两个参数：一个字节类型的参数label和一个指向node类型的指针node。它没有返回值。
func (n *node) updateEdge(label byte, node *node) {
	// 它获取当前节点边的数量num
	num := len(n.edges)
	// 使用sort.Search函数在边中查找给定的标签。
	// sort.Search函数接受两个参数：一个是要搜索的元素的数量，另一个是一个函数，这个函数接受一个索引并返回一个布尔值，表示是否找到了要搜索的元素。
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	// 找到了给定的标签，更新对应的节点。否则，抛出一个panic。
	if idx < num && n.edges[idx].label == label {
		n.edges[idx].node = node
		return
	}
	panic("replacing missing edge")
}

// 获取当前节点的一个边
func (n *node) getEdge(label byte) *node {
	// 获取当前节点边的数量num
	num := len(n.edges)
	// 使用sort.Search函数在边中查找给定的标签。
	// sort.Search函数接受两个参数：一个是要搜索的元素的数量，另一个是一个函数，这个函数接受一个索引并返回一个布尔值，表示是否找到了要搜索的元素。
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	// 如果找到了给定的标签，返回对应的节点。否则，返回nil。
	if idx < num && n.edges[idx].label == label {
		return n.edges[idx].node
	}
	return nil
}

// 用于删除当前节点的一个边
// 方法接受一个字节类型的参数label。它没有返回值。
func (n *node) delEdge(label byte) {
	// 它获取当前节点边的数量num
	num := len(n.edges)
	// 使用sort.Search函数在边中查找给定的标签。
	// sort.Search函数接受两个参数：一个是要搜索的元素的数量，另一个是一个函数，这个函数接受一个索引并返回一个布尔值，表示是否找到了要搜索的元素。
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	// 如果找到了给定的标签，删除对应的边。删除操作是通过将从找到的索引开始的所有边向前移动一位，然后将最后一个边设置为零值，并将边的切片缩短一位来实现的。
	if idx < num && n.edges[idx].label == label {
		copy(n.edges[idx:], n.edges[idx+1:])
		n.edges[len(n.edges)-1] = edge{}
		n.edges = n.edges[:len(n.edges)-1]
	}
}

// 这是一个名为`edges`的类型定义，它是`edge`类型的切片。这个类型定义了一些方法，使得它可以满足Go语言的`sort.Interface`接口，从而可以使用`sort`包的排序函数。
// 这个类型定义了四个方法：
// 1. `Len`：这个方法返回边的数量，它满足了`sort.Interface`接口的`Len`方法。
// 2. `Less`：这个方法接受两个索引`i`和`j`，返回一个布尔值，表示在索引`i`处的边的标签是否小于在索引`j`处的边的标签。它满足了`sort.Interface`接口的`Less`方法。
// 3. `Swap`：这个方法接受两个索引`i`和`j`，并交换这两个索引处的边。它满足了`sort.Interface`接口的`Swap`方法。
// 4. `Sort`：这个方法使用`sort.Sort`函数对边进行排序。
// 这个类型的主要作用是在前缀树（Trie）的节点中存储边，并提供了排序和比较的功能。
type edges []edge

func (e edges) Len() int {
	return len(e)
}

func (e edges) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges) Sort() {
	sort.Sort(e)
}

// Tree implements a radix tree. This can be treated as a
// Dictionary abstract data type. The main advantage over
// a standard hash map is prefix-based lookups and
// ordered iteration,
//
// 实现一个基数树（Radix Tree）。基数树可以被视为一种字典抽象数据类型。与标准哈希映射相比，它的主要优势在于基于前缀的查找和有序迭代。
type Tree struct {
	root *node // 指向node类型的指针，表示树的根节点。
	size int   // 表示树的大小，即树中节点的数量。
}

// New returns an empty Tree
func New() *Tree {
	return NewFromMap(nil)
}

// NewFromMap returns a new tree containing the keys
// from an existing map
func NewFromMap(m map[string]interface{}) *Tree {
	t := &Tree{
		root: &node{},
	}
	for k, v := range m {
		t.Insert(k, v)
	}
	return t
}

// Len is used to return the number of elements in the tree
func (t *Tree) Len() int {
	return t.size
}

// longestPrefix finds the length of the shared prefix
// of two strings
//
// 找出两个字符串的最长公共前缀的长度
func longestPrefix(k1, k2 string) int {
	// 找出两个字符串中较短的那个的长度max
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}
	// 使用一个循环来比较两个字符串的每一个字符。如果在某个位置上，两个字符串的字符不相同，它就会跳出循环。
	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	// 返回循环的次数，即最长公共前缀的长度。
	return i
}

// Insert is used to add a newentry or update
// an existing entry. Returns true if an existing record is updated.
//
// 用于在一个前缀树（Trie）中插入或更新一个键值对
// 方法接受两个参数：一个字符串s和一个空接口类型的值v。
// 它返回两个值：一个空接口类型的值和一个布尔值。如果更新了现有的记录，返回true。
func (t *Tree) Insert(s string, v interface{}) (interface{}, bool) {
	// 初始化两个节点类型的变量parent和n，并将搜索的键设置为s。
	var parent *node // 这是一个指向node类型的指针，用于存储当前节点的父节点。在遍历树的过程中，parent始终指向当前节点的父节点。
	n := t.root      // 这是一个指向node类型的指针，用于存储当前节点。在开始时，n被设置为树的根节点。在遍历树的过程中，n始终指向当前正在处理的节点。
	search := s      // 用于存储正在搜索的键。在开始时，search被设置为要插入的键s。在遍历树的过程中，search会被更新为剩余的未匹配的键。

	// 进入一个无限循环
	for {
		// 首先检查搜索的键是否已经用完。如果用完了，检查当前节点是否是叶子节点。
		// 如果是，更新叶子节点的值并返回旧值和true。否则，创建一个新的叶子节点，并增加树的大小。
		// Handle key exhaution
		if len(search) == 0 {
			// 叶子节点，修改老值返回
			if n.isLeaf() {
				old := n.leaf.val
				n.leaf.val = v
				return old, true
			}
			// 创建新的的叶子节点插入
			n.leaf = &leafNode{
				key: s,
				val: v,
			}
			t.size++
			return nil, false
		}

		// 查找边
		// Look for the edge
		parent = n
		n = n.getEdge(search[0])

		// 没有找到，创建一个新的边，并增加树的大小。
		// No edge, create one
		if n == nil {
			e := edge{ // 边节点
				label: search[0], // 边的标签
				node: &node{ // 节点
					leaf: &leafNode{ // 叶子节点
						key: s,
						val: v,
					},
					prefix: search, // 公共前缀
				},
			}
			// 在当前节点添加一个边
			parent.addEdge(e)
			t.size++
			return nil, false
		}

		// 找到了边，计算搜索的键和节点前缀的最长公共前缀。如果公共前缀的长度等于节点前缀的长度，更新搜索的键并继续循环。
		// Determine longest prefix of the search key on match
		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// 如果公共前缀的长度小于节点前缀的长度，分裂节点，并创建一个新的叶子节点。
		// 然后，根据搜索的键的长度来决定是将新的叶子节点添加到当前节点，还是创建一个新的边。
		// Split the node
		t.size++
		// 创建一个新的节点child，其前缀是键的公共前缀部分
		child := &node{
			prefix: search[:commonPrefix],
		}
		// 更新父节点parent的边，使其指向新创建的child节点
		parent.updateEdge(search[0], child)

		// 将原节点n作为child的一个子节点，更新n的前缀为原前缀去掉公共前缀部分
		// Restore the existing node
		child.addEdge(edge{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		// 创建一个新的叶子节点leaf，其键和值分别为插入的键和值
		// Create a new leaf node
		leaf := &leafNode{
			key: s,
			val: v,
		}

		// 如果新的键是原键的子集（即公共前缀后无剩余字符），则将leaf作为child的叶子节点。
		// If the new key is a subset, add to this node
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.leaf = leaf
			return nil, false
		}

		// 如果新的键不是原键的子集（即公共前缀后还有剩余字符），则创建一个新的边，其标签为剩余键的第一个字符，节点为一个新的节点，其叶子节点为leaf，前缀为剩余的键。
		// Create a new edge for the node
		child.addEdge(edge{
			label: search[0],
			node: &node{
				leaf:   leaf,
				prefix: search,
			},
		})
		return nil, false
	}

}

// Delete is used to delete a key, returning the previous
// value and if it was deleted
func (t *Tree) Delete(s string) (interface{}, bool) {
	var parent *node
	var label byte
	n := t.root
	search := s
	for {
		// Check for key exhaution
		if len(search) == 0 {
			if !n.isLeaf() {
				break
			}
			goto DELETE
		}

		// Look for an edge
		parent = n
		label = search[0]
		n = n.getEdge(label)
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return nil, false

DELETE:
	// Delete the leaf
	leaf := n.leaf
	n.leaf = nil
	t.size--

	// Check if we should delete this node from the parent
	if parent != nil && len(n.edges) == 0 {
		parent.delEdge(label)
	}

	// Check if we should merge this node
	if n != t.root && len(n.edges) == 1 {
		n.mergeChild()
	}

	// Check if we should merge the parent's other child
	if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
		parent.mergeChild()
	}

	return leaf.val, true
}

// DeletePrefix is used to delete the subtree under a prefix
// Returns how many nodes were deleted
// Use this to delete large subtrees efficiently
func (t *Tree) DeletePrefix(s string) int {
	return t.deletePrefix(nil, t.root, s)
}

// delete does a recursive deletion
func (t *Tree) deletePrefix(parent, n *node, prefix string) int {
	// Check for key exhaustion
	if len(prefix) == 0 {
		// Remove the leaf node
		subTreeSize := 0
		//recursively walk from all edges of the node to be deleted
		recursiveWalk(n, func(s string, v interface{}) bool {
			subTreeSize++
			return false
		})
		if n.isLeaf() {
			n.leaf = nil
		}
		n.edges = nil // deletes the entire subtree

		// Check if we should merge the parent's other child
		if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
			parent.mergeChild()
		}
		t.size -= subTreeSize
		return subTreeSize
	}

	// Look for an edge
	label := prefix[0]
	child := n.getEdge(label)
	if child == nil || (!strings.HasPrefix(child.prefix, prefix) && !strings.HasPrefix(prefix, child.prefix)) {
		return 0
	}

	// Consume the search prefix
	if len(child.prefix) > len(prefix) {
		prefix = prefix[len(prefix):]
	} else {
		prefix = prefix[len(child.prefix):]
	}
	return t.deletePrefix(n, child, prefix)
}

func (n *node) mergeChild() {
	e := n.edges[0]
	child := e.node
	n.prefix = n.prefix + child.prefix
	n.leaf = child.leaf
	n.edges = child.edges
}

// Get is used to lookup a specific key, returning
// the value and if it was found
func (t *Tree) Get(s string) (interface{}, bool) {
	n := t.root
	search := s
	for {
		// Check for key exhaution
		if len(search) == 0 {
			if n.isLeaf() {
				return n.leaf.val, true
			}
			break
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return nil, false
}

// LongestPrefix is like Get, but instead of an
// exact match, it will return the longest prefix match.
func (t *Tree) LongestPrefix(s string) (string, interface{}, bool) {
	var last *leafNode
	n := t.root
	search := s
	for {
		// Look for a leaf node
		if n.isLeaf() {
			last = n.leaf
		}

		// Check for key exhaution
		if len(search) == 0 {
			break
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	if last != nil {
		return last.key, last.val, true
	}
	return "", nil, false
}

// Minimum is used to return the minimum value in the tree
func (t *Tree) Minimum() (string, interface{}, bool) {
	n := t.root
	for {
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		if len(n.edges) > 0 {
			n = n.edges[0].node
		} else {
			break
		}
	}
	return "", nil, false
}

// Maximum is used to return the maximum value in the tree
func (t *Tree) Maximum() (string, interface{}, bool) {
	n := t.root
	for {
		if num := len(n.edges); num > 0 {
			n = n.edges[num-1].node
			continue
		}
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		break
	}
	return "", nil, false
}

// Walk is used to walk the tree
func (t *Tree) Walk(fn WalkFn) {
	recursiveWalk(t.root, fn)
}

// WalkPrefix is used to walk the tree under a prefix
func (t *Tree) WalkPrefix(prefix string, fn WalkFn) {
	n := t.root
	search := prefix
	for {
		// Check for key exhaustion
		if len(search) == 0 {
			recursiveWalk(n, fn)
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
			continue
		}
		if strings.HasPrefix(n.prefix, search) {
			// Child may be under our search prefix
			recursiveWalk(n, fn)
		}
		return
	}
}

// WalkPath is used to walk the tree, but only visiting nodes
// from the root down to a given leaf. Where WalkPrefix walks
// all the entries *under* the given prefix, this walks the
// entries *above* the given prefix.
func (t *Tree) WalkPath(path string, fn WalkFn) {
	n := t.root
	search := path
	for {
		// Visit the leaf values if any
		if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
			return
		}

		// Check for key exhaution
		if len(search) == 0 {
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if strings.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
}

// recursiveWalk is used to do a pre-order walk of a node
// recursively. Returns true if the walk should be aborted
func recursiveWalk(n *node, fn WalkFn) bool {
	// Visit the leaf values if any
	if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
		return true
	}

	// Recurse on the children
	i := 0
	k := len(n.edges) // keeps track of number of edges in previous iteration
	for i < k {
		e := n.edges[i]
		if recursiveWalk(e.node, fn) {
			return true
		}
		// It is a possibility that the WalkFn modified the node we are
		// iterating on. If there are no more edges, mergeChild happened,
		// so the last edge became the current node n, on which we'll
		// iterate one last time.
		if len(n.edges) == 0 {
			return recursiveWalk(n, fn)
		}
		// If there are now less edges than in the previous iteration,
		// then do not increment the loop index, since the current index
		// points to a new edge. Otherwise, get to the next index.
		if len(n.edges) >= k {
			i++
		}
		k = len(n.edges)
	}
	return false
}

// ToMap is used to walk the tree and convert it into a map
func (t *Tree) ToMap() map[string]interface{} {
	out := make(map[string]interface{}, t.size)
	t.Walk(func(k string, v interface{}) bool {
		out[k] = v
		return false
	})
	return out
}
