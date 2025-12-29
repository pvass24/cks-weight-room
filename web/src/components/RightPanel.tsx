'use client'

import { useState } from 'react'
import MultiTerminal from './MultiTerminal'
import IDEView from './IDEView'

type View = 'terminal' | 'ide'

interface RightPanelProps {
  exerciseSlug: string
  clusterReady: boolean
  selectedNode: string
  onNodeChange?: (nodeName: string) => void
}

export default function RightPanel({
  exerciseSlug,
  clusterReady,
  selectedNode,
  onNodeChange
}: RightPanelProps) {
  const [activeView, setActiveView] = useState<View>('terminal')

  return (
    <div className="flex-1 flex flex-col bg-gray-900">
      {/* View Switcher Tabs */}
      <div className="bg-gray-800 border-b border-gray-700 flex items-center px-4">
        <button
          onClick={() => setActiveView('terminal')}
          className={`px-4 py-3 text-sm font-medium border-b-2 transition ${
            activeView === 'terminal'
              ? 'border-blue-500 text-white'
              : 'border-transparent text-gray-400 hover:text-gray-200'
          }`}
        >
          <div className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            Terminal
          </div>
        </button>
        <button
          onClick={() => setActiveView('ide')}
          className={`px-4 py-3 text-sm font-medium border-b-2 transition ${
            activeView === 'ide'
              ? 'border-blue-500 text-white'
              : 'border-transparent text-gray-400 hover:text-gray-200'
          }`}
        >
          <div className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
            IDE
          </div>
        </button>
      </div>

      {/* View Content */}
      {activeView === 'terminal' ? (
        <MultiTerminal
          exerciseSlug={exerciseSlug}
          clusterReady={clusterReady}
          selectedNode={selectedNode}
          onNodeChange={onNodeChange}
        />
      ) : (
        <IDEView
          exerciseSlug={exerciseSlug}
          clusterReady={clusterReady}
          selectedNode={selectedNode}
        />
      )}
    </div>
  )
}
