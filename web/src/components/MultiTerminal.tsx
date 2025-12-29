'use client'

import { useState, useEffect, useRef } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { Plus, X, Maximize2, Columns, Rows } from 'lucide-react'
import '@xterm/xterm/css/xterm.css'

interface Node {
  name: string
  role: string
}

interface TerminalInstance {
  id: string
  term: XTerm
  ws: WebSocket
  fitAddon: FitAddon
  title: string
}

interface TerminalPane {
  id: string
  terminals: TerminalInstance[]
  activeTerminalId: string
}

type SplitDirection = 'horizontal' | 'vertical' | null

interface MultiTerminalProps {
  exerciseSlug: string
  clusterReady: boolean
  selectedNode?: string
  onNodeChange?: (nodeName: string) => void
}

export default function MultiTerminal({
  exerciseSlug,
  clusterReady,
  selectedNode: controlledSelectedNode,
  onNodeChange
}: MultiTerminalProps) {
  const [panes, setPanes] = useState<TerminalPane[]>([])
  const [splitDirection, setSplitDirection] = useState<SplitDirection>(null)
  const [nodes, setNodes] = useState<Node[]>([])
  const [internalSelectedNode, setInternalSelectedNode] = useState<string>('')

  // Use controlled prop if provided, otherwise use internal state
  const selectedNode = controlledSelectedNode !== undefined ? controlledSelectedNode : internalSelectedNode
  const setSelectedNode = (nodeName: string) => {
    setInternalSelectedNode(nodeName)
    onNodeChange?.(nodeName)
  }
  const terminalRefs = useRef<Map<string, HTMLDivElement>>(new Map())
  const instancesRef = useRef<Map<string, TerminalInstance>>(new Map())
  const isInitialMount = useRef(true)

  // Fetch nodes when cluster is ready
  useEffect(() => {
    if (!clusterReady) return

    const fetchNodes = async () => {
      try {
        const response = await fetch(`/api/cluster/nodes/${exerciseSlug}`)
        const data = await response.json()
        console.log('Fetched nodes:', data.nodes)
        if (data.success && data.nodes) {
          setNodes(data.nodes)
          // Select control plane by default
          const controlPlane = data.nodes.find((n: Node) => n.role === 'control-plane')
          console.log('Control plane node:', controlPlane)
          if (controlPlane) {
            console.log('Setting selectedNode to:', controlPlane.name)
            setSelectedNode(controlPlane.name)
          } else {
            console.warn('No control plane node found!')
          }
        }
      } catch (err) {
        console.error('Failed to fetch nodes:', err)
      }
    }

    fetchNodes()
  }, [clusterReady, exerciseSlug])

  // Initialize first terminal AFTER nodes are fetched
  useEffect(() => {
    if (!clusterReady || panes.length > 0 || !selectedNode) return

    console.log('Initializing first terminal for node:', selectedNode)

    const paneId = 'pane-1'
    const terminalId = `${paneId}-terminal-1`

    // Create placeholder terminal entry first so the div gets rendered
    const placeholderInstance: TerminalInstance = {
      id: terminalId,
      term: null as any, // Will be initialized below
      ws: null as any,
      fitAddon: null as any,
      title: 'Terminal 1'
    }

    setPanes([{
      id: paneId,
      terminals: [placeholderInstance],
      activeTerminalId: terminalId
    }])

    // Small delay to ensure DOM is ready, then initialize the terminal
    setTimeout(() => {
      initializeTerminal(paneId, terminalId, 'Terminal 1', selectedNode)
    }, 100)
  }, [clusterReady, selectedNode])

  // Reconnect all terminals when node selection changes
  useEffect(() => {
    // Skip on initial mount
    if (isInitialMount.current) {
      isInitialMount.current = false
      return
    }

    if (!selectedNode || panes.length === 0) return

    console.log('Node changed to:', selectedNode, '- reconnecting terminals')

    // Close all existing terminals and reconnect to new node
    panes.forEach(pane => {
      pane.terminals.forEach(terminal => {
        const instance = instancesRef.current.get(terminal.id)
        if (instance) {
          // Close existing connection
          instance.ws.close()
          instance.term.dispose()
          instancesRef.current.delete(terminal.id)
        }
      })
    })

    // Reinitialize all terminals with new node
    setTimeout(() => {
      panes.forEach(pane => {
        pane.terminals.forEach(terminal => {
          initializeTerminal(pane.id, terminal.id, terminal.title, selectedNode)
        })
      })
    }, 100)
  }, [selectedNode])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      instancesRef.current.forEach(instance => {
        instance.ws.close()
        instance.term.dispose()
      })
      instancesRef.current.clear()
    }
  }, [])

  const initializeTerminal = (paneId: string, terminalId: string, title: string, targetNode?: string) => {
    const terminalEl = terminalRefs.current.get(terminalId)
    if (!terminalEl) {
      console.error(`Terminal element not found for ${terminalId}`)
      return
    }

    // Use provided node or fall back to currently selected node
    const nodeForTerminal = targetNode || selectedNode
    console.log(`Initializing terminal ${terminalId} for node: ${nodeForTerminal}`)

    // Create terminal instance
    const term = new XTerm({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#e5e5e5',
      },
    })

    // Add addons
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.loadAddon(new WebLinksAddon())

    // Open terminal in DOM
    term.open(terminalEl)
    setTimeout(() => {
      fitAddon.fit()
      term.focus() // Give terminal focus to receive keyboard input
    }, 50)

    // Connect WebSocket (include target node if specified)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    let wsUrl = `${protocol}//${window.location.host}/api/terminal/${exerciseSlug}`
    if (nodeForTerminal) {
      wsUrl += `?node=${encodeURIComponent(nodeForTerminal)}`
    }
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      term.writeln('\x1b[32m✓ Connected to terminal\x1b[0m')
      term.writeln('')

      // Send terminal size
      ws.send(JSON.stringify({
        type: 'resize',
        rows: term.rows,
        cols: term.cols,
      }))
    }

    ws.onmessage = (event) => {
      term.write(event.data)
    }

    ws.onerror = () => {
      term.writeln('\x1b[31m✗ WebSocket error\x1b[0m')
    }

    ws.onclose = () => {
      term.writeln('')
      term.writeln('\x1b[33m✗ Connection closed\x1b[0m')
    }

    // Handle terminal input
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        // Intercept Ctrl+L (form feed, 0x0C) and send clear command instead
        if (data === '\f') {
          ws.send(JSON.stringify({
            type: 'input',
            data: 'clear\n',
          }))
        } else {
          ws.send(JSON.stringify({
            type: 'input',
            data: data,
          }))
        }
      }
    })

    // Handle terminal resize
    term.onResize(({ rows, cols }) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          rows,
          cols,
        }))
      }
    })

    // Store instance
    const instance: TerminalInstance = {
      id: terminalId,
      term,
      ws,
      fitAddon,
      title
    }
    instancesRef.current.set(terminalId, instance)

    // Update the placeholder terminal with the real instance
    setPanes(prev => prev.map(pane => {
      if (pane.id === paneId) {
        return {
          ...pane,
          terminals: pane.terminals.map(t =>
            t.id === terminalId ? instance : t
          )
        }
      }
      return pane
    }))
  }

  const addTab = (paneId: string) => {
    const pane = panes.find(p => p.id === paneId)
    if (!pane) return

    const terminalId = `${paneId}-terminal-${pane.terminals.length + 1}`
    const title = `Terminal ${pane.terminals.length + 1}`

    // Create placeholder terminal entry first
    const placeholderInstance: TerminalInstance = {
      id: terminalId,
      term: null as any,
      ws: null as any,
      fitAddon: null as any,
      title
    }

    setPanes(prev => prev.map(p => {
      if (p.id === paneId) {
        return {
          ...p,
          terminals: [...p.terminals, placeholderInstance],
          activeTerminalId: terminalId
        }
      }
      return p
    }))

    // Small delay to ensure DOM is ready
    setTimeout(() => {
      initializeTerminal(paneId, terminalId, title, selectedNode)
    }, 100)
  }

  const closeTab = (paneId: string, terminalId: string) => {
    const instance = instancesRef.current.get(terminalId)
    if (instance) {
      instance.ws.close()
      instance.term.dispose()
      instancesRef.current.delete(terminalId)
    }

    setPanes(prev => prev.map(pane => {
      if (pane.id === paneId) {
        const newTerminals = pane.terminals.filter(t => t.id !== terminalId)
        const newActiveId = newTerminals.length > 0
          ? (pane.activeTerminalId === terminalId ? newTerminals[0].id : pane.activeTerminalId)
          : ''
        return {
          ...pane,
          terminals: newTerminals,
          activeTerminalId: newActiveId
        }
      }
      return pane
    }))
  }

  const splitPane = (direction: 'horizontal' | 'vertical') => {
    const newPaneId = `pane-${panes.length + 1}`
    const newTerminalId = `${newPaneId}-terminal-1`

    // Create placeholder terminal entry first
    const placeholderInstance: TerminalInstance = {
      id: newTerminalId,
      term: null as any,
      ws: null as any,
      fitAddon: null as any,
      title: 'Terminal 1'
    }

    setSplitDirection(direction)
    setPanes(prev => [...prev, {
      id: newPaneId,
      terminals: [placeholderInstance],
      activeTerminalId: newTerminalId
    }])

    // Small delay to ensure DOM is ready
    setTimeout(() => {
      initializeTerminal(newPaneId, newTerminalId, 'Terminal 1', selectedNode)
    }, 100)
  }

  const closePane = (paneId: string) => {
    const pane = panes.find(p => p.id === paneId)
    if (!pane) return

    // Close all terminals in pane
    pane.terminals.forEach(terminal => {
      const instance = instancesRef.current.get(terminal.id)
      if (instance) {
        instance.ws.close()
        instance.term.dispose()
        instancesRef.current.delete(terminal.id)
      }
    })

    const newPanes = panes.filter(p => p.id !== paneId)
    setPanes(newPanes)

    // Reset split direction if only one pane left
    if (newPanes.length === 1) {
      setSplitDirection(null)
    }
  }

  if (!clusterReady) {
    return (
      <div className="flex-1 bg-gray-900 flex items-center justify-center">
        <div className="text-center text-gray-400">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-500 mx-auto mb-4"></div>
          <p className="text-lg">Waiting for cluster to be ready...</p>
        </div>
      </div>
    )
  }

  const containerClass = splitDirection === 'horizontal'
    ? 'flex flex-col'
    : splitDirection === 'vertical'
    ? 'flex flex-row'
    : 'flex flex-col'

  return (
    <div className="flex-1 bg-gray-900 flex flex-col">
      {/* Main toolbar */}
      <div className="bg-gray-800 px-4 py-2 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-3 h-3 rounded-full bg-red-500"></div>
            <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
            <div className="w-3 h-3 rounded-full bg-green-500"></div>
          </div>
          <span className="text-sm text-gray-400 ml-2">Multi-Terminal</span>
        </div>

        {/* Node selector */}
        {nodes.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-500">Node:</span>
            <select
              value={selectedNode}
              onChange={(e) => setSelectedNode(e.target.value)}
              className="bg-gray-700 text-gray-200 text-sm px-3 py-1 rounded border border-gray-600 focus:outline-none focus:border-blue-500"
            >
              {nodes.map((node) => (
                <option key={node.name} value={node.name}>
                  {node.name} ({node.role})
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="flex items-center gap-2">
          <button
            onClick={() => splitPane('vertical')}
            className="text-gray-400 hover:text-white p-1 hover:bg-gray-700 rounded"
            title="Split Vertical"
          >
            <Columns className="w-4 h-4" />
          </button>
          <button
            onClick={() => splitPane('horizontal')}
            className="text-gray-400 hover:text-white p-1 hover:bg-gray-700 rounded"
            title="Split Horizontal"
          >
            <Rows className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Terminal panes */}
      <div className={`flex-1 ${containerClass}`}>
        {panes.map((pane, paneIndex) => (
          <div key={pane.id} className="flex-1 flex flex-col border-gray-700" style={{
            borderLeftWidth: splitDirection === 'vertical' && paneIndex > 0 ? '1px' : '0',
            borderTopWidth: splitDirection === 'horizontal' && paneIndex > 0 ? '1px' : '0'
          }}>
            {/* Pane toolbar with tabs */}
            <div className="bg-gray-800 border-b border-gray-700 flex items-center justify-between">
              <div className="flex items-center overflow-x-auto">
                {pane.terminals.map((terminal) => (
                  <div
                    key={terminal.id}
                    className={`flex items-center gap-2 px-3 py-2 text-sm cursor-pointer border-r border-gray-700 ${
                      terminal.id === pane.activeTerminalId
                        ? 'bg-gray-900 text-white'
                        : 'bg-gray-800 text-gray-400 hover:bg-gray-750'
                    }`}
                    onClick={() => {
                      setPanes(prev => prev.map(p =>
                        p.id === pane.id ? { ...p, activeTerminalId: terminal.id } : p
                      ))
                    }}
                  >
                    <span>{terminal.title}</span>
                    {pane.terminals.length > 1 && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          closeTab(pane.id, terminal.id)
                        }}
                        className="hover:text-red-400"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    )}
                  </div>
                ))}
                <button
                  onClick={() => addTab(pane.id)}
                  className="px-3 py-2 text-gray-400 hover:text-white hover:bg-gray-700"
                  title="New Tab"
                >
                  <Plus className="w-4 h-4" />
                </button>
              </div>
              {panes.length > 1 && (
                <button
                  onClick={() => closePane(pane.id)}
                  className="px-3 py-2 text-gray-400 hover:text-red-400"
                  title="Close Pane"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>

            {/* Terminal instances */}
            <div className="flex-1 relative">
              {pane.terminals.map((terminal) => (
                <div
                  key={terminal.id}
                  ref={(el) => {
                    if (el) terminalRefs.current.set(terminal.id, el)
                  }}
                  className={`absolute inset-0 p-2 ${
                    terminal.id === pane.activeTerminalId ? 'block' : 'hidden'
                  }`}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
