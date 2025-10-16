package prompt

import (
	"strings"
	"testing"

	"github.com/agentflare-ai/go-xmldom"
)

func TestPruneSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Elements that should exist
		removed  []string // Elements that should be removed
	}{
		{
			name: "Remove action elements but preserve structure",
			input: `<?xml version="1.0"?>
<scxml xmlns="http://www.w3.org/2005/07/scxml" version="1.0" initial="s1">
	<datamodel>
		<data id="x" expr="1"/>
	</datamodel>
	<state id="s1">
		<onentry>
			<log expr="'entering s1'"/>
			<assign location="x" expr="2"/>
		</onentry>
		<transition event="e" target="s2">
			<log expr="'taking transition'"/>
			<send event="foo"/>
		</transition>
		<onexit>
			<log expr="'leaving s1'"/>
		</onexit>
	</state>
	<parallel id="p1">
		<state id="s2">
			<initial>
				<transition target="s2_1"/>
			</initial>
			<state id="s2_1">
				<transition event="done" target="s2_final"/>
			</state>
			<final id="s2_final"/>
		</state>
		<history id="h1" type="deep"/>
	</parallel>
	<final id="done">
		<donedata>
			<param name="result" expr="'completed'"/>
		</donedata>
	</final>
</scxml>`,
			expected: []string{
				"scxml",
				"state[@id='s1']",
				"transition[@event='e']",
				"parallel[@id='p1']",
				"state[@id='s2']",
				"initial",
				"state[@id='s2_1']",
				"final[@id='s2_final']",
				"history[@id='h1']",
				"final[@id='done']",
			},
			removed: []string{
				"datamodel",
				"data",
				"onentry",
				"onexit",
				"log",
				"assign",
				"send",
				"donedata",
				"param",
			},
		},
		{
			name: "Preserve runtime:datamodel but remove static datamodel",
			input: `<?xml version="1.0"?>
<scxml xmlns="http://www.w3.org/2005/07/scxml" 
       xmlns:runtime="urn:gogo:scxml:runtime:1"
       version="1.0">
	<datamodel>
		<data id="static" expr="1"/>
	</datamodel>
	<runtime:datamodel>
		<runtime:data id="dynamic" value="2"/>
	</runtime:datamodel>
	<state id="s1">
		<transition event="e" target="s2"/>
	</state>
	<state id="s2"/>
</scxml>`,
			expected: []string{
				"scxml",
				"runtime:datamodel",
				"runtime:data[@id='dynamic']",
				"state[@id='s1']",
				"state[@id='s2']",
				"transition[@event='e']",
			},
			removed: []string{
				"scxml:datamodel", // The regular SCXML datamodel element should be removed
				"scxml:data",      // The static data element should be removed with its parent
			},
		},
		{
			name: "Remove executable content from transitions",
			input: `<?xml version="1.0"?>
<scxml xmlns="http://www.w3.org/2005/07/scxml" version="1.0">
	<state id="s1">
		<transition event="test">
			<log expr="'transition log'"/>
			<script>console.log('test');</script>
			<if cond="x > 0">
				<assign location="y" expr="1"/>
			<elseif cond="x == 0"/>
				<assign location="y" expr="0"/>
			<else/>
				<assign location="y" expr="-1"/>
			</if>
			<foreach array="items" item="i">
				<log expr="i"/>
			</foreach>
		</transition>
	</state>
</scxml>`,
			expected: []string{
				"scxml",
				"state[@id='s1']",
				"transition[@event='test']",
			},
			removed: []string{
				"log",
				"script",
				"if",
				"elseif",
				"else",
				"assign",
				"foreach",
			},
		},
		{
			name: "Keep invoke but remove finalize",
			input: `<?xml version="1.0"?>
<scxml xmlns="http://www.w3.org/2005/07/scxml" version="1.0">
	<state id="s1">
		<invoke id="inv1" type="scxml" src="child.scxml">
			<param name="x" expr="1"/>
			<finalize>
				<log expr="'finalize'"/>
			</finalize>
		</invoke>
	</state>
</scxml>`,
			expected: []string{
				"scxml",
				"state[@id='s1']",
				"invoke[@id='inv1']",
			},
			removed: []string{
				"param",
				"finalize",
				"log",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input XML
			decoder := xmldom.NewDecoder(strings.NewReader(tt.input))
			doc, err := decoder.Decode()
			if err != nil {
				t.Fatalf("Failed to parse XML: %v", err)
			}

			// Apply PruneSnapshot
			PruneSnapshot(doc)

			// Check that expected elements exist
			for _, xpath := range tt.expected {
				// Simple check - just verify the element exists
				if !elementExists(doc.DocumentElement(), xpath) {
					t.Errorf("Expected element %q to exist but it was not found", xpath)
				}
			}

			// Check that removed elements don't exist
			for _, xpath := range tt.removed {
				if elementExists(doc.DocumentElement(), xpath) {
					t.Errorf("Expected element %q to be removed but it still exists", xpath)
				}
			}
		})
	}
}

// elementExists is a simple helper to check if an element matching the pattern exists
func elementExists(root xmldom.Element, pattern string) bool {
	if root == nil {
		return false
	}

	// Simple pattern matching (not full XPath)
	parts := strings.Split(pattern, "[@")
	tagName := parts[0]

	// Handle namespace prefixes
	var namespace, localName string
	if idx := strings.Index(tagName, ":"); idx > 0 {
		namespace = tagName[:idx]
		localName = tagName[idx+1:]
	} else {
		localName = tagName
	}

	// Check if root matches
	if string(root.LocalName()) == localName {
		if namespace != "" {
			// Check namespace
			if namespace == "runtime" && root.NamespaceURI() == "urn:gogo:scxml:runtime:1" {
				return true
			} else if namespace == "scxml" && root.NamespaceURI() == "http://www.w3.org/2005/07/scxml" {
				return true
			}
		} else {
			// No namespace prefix specified - match elements without considering namespace
			return true
		}
	}

	// Check children recursively
	return elementExistsRecursive(root, localName, namespace)
}

func elementExistsRecursive(elem xmldom.Element, localName, namespace string) bool {
	children := elem.ChildNodes()
	for i := uint(0); i < children.Length(); i++ {
		if child, ok := children.Item(i).(xmldom.Element); ok {
			childLocalName := string(child.LocalName())
			if childLocalName == localName {
				if namespace != "" {
					if namespace == "runtime" && child.NamespaceURI() == "urn:gogo:scxml:runtime:1" {
						return true
					} else if namespace == "scxml" && child.NamespaceURI() == "http://www.w3.org/2005/07/scxml" {
						return true
					}
				} else {
					// No namespace prefix - match any
					return true
				}
			}
			// Recurse
			if elementExistsRecursive(child, localName, namespace) {
				return true
			}
		}
	}
	return false
}
