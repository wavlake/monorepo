import { useState } from 'react'

// Temporarily define event kinds locally until workspace is set up
enum NostrEventKind {
  TEXT_NOTE = 1,
  TRACK_METADATA = 31337,
}

function App() {
  const [count, setCount] = useState(0)

  // Example of using shared types
  const handleNostrExample = () => {
    console.log('Nostr Text Note Kind:', NostrEventKind.TEXT_NOTE)
    console.log('Nostr Track Metadata Kind:', NostrEventKind.TRACK_METADATA)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <header className="text-center mb-12">
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 backdrop-blur-sm border border-gray-200 dark:border-gray-700">
            <h1 className="text-5xl font-bold mb-4 bg-gradient-to-r from-primary-600 to-secondary-600 bg-clip-text text-transparent">
              🎵 Wavlake
            </h1>
            <p className="text-xl text-gray-600 dark:text-gray-300 mb-2">
              Decentralized Music Platform
            </p>
            <p className="text-gray-500 dark:text-gray-400">
              Built with React + TypeScript + Vite + Go + Nostr + Tailwind
            </p>
          </div>
        </header>

        {/* Interactive Demo */}
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 mb-8 border border-gray-200 dark:border-gray-700">
          <h2 className="text-2xl font-semibold mb-4 text-gray-900 dark:text-white">
            Interactive Demo
          </h2>
          <div className="flex gap-4 justify-center">
            <button 
              onClick={() => setCount((count) => count + 1)}
              className="bg-primary-500 hover:bg-primary-600 text-white font-medium py-3 px-6 rounded-lg transition-colors duration-200 shadow-md hover:shadow-lg"
            >
              Count is {count}
            </button>
            <button 
              onClick={handleNostrExample}
              className="bg-secondary-500 hover:bg-secondary-600 text-white font-medium py-3 px-6 rounded-lg transition-colors duration-200 shadow-md hover:shadow-lg"
            >
              Test Shared Types
            </button>
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-2 gap-8 mb-8">
          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 border border-gray-200 dark:border-gray-700">
            <h2 className="text-2xl font-semibold mb-4 text-gray-900 dark:text-white flex items-center">
              🚀 Features
            </h2>
            <ul className="space-y-3">
              {[
                'React 18 + TypeScript',
                'Vite for fast development', 
                'Tailwind CSS for styling',
                'Shared types from Go backend',
                'Nostr integration ready',
                'TDD setup with Vitest',
                'Firebase Auth ready',
                'Component testing with React Testing Library',
                'E2E testing with Playwright'
              ].map((feature, index) => (
                <li key={index} className="flex items-center text-gray-700 dark:text-gray-300">
                  <span className="text-green-500 mr-3">✅</span>
                  {feature}
                </li>
              ))}
            </ul>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 border border-gray-200 dark:border-gray-700">
            <h2 className="text-2xl font-semibold mb-4 text-gray-900 dark:text-white flex items-center">
              🧪 Development Commands
            </h2>
            <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-4 font-mono text-sm overflow-x-auto">
              <div className="text-gray-800 dark:text-gray-200 space-y-2">
                <div className="text-green-600 dark:text-green-400"># TDD workflow</div>
                <div>task tdd:frontend    <span className="text-gray-500"># Start test watcher</span></div>
                <div>task dev:frontend    <span className="text-gray-500"># Start dev server</span></div>
                <div>task test:unit:frontend  <span className="text-gray-500"># Run tests</span></div>
                <div>task build:frontend  <span className="text-gray-500"># Build for production</span></div>
              </div>
            </div>
          </div>
        </div>

        {/* Tech Stack */}
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 border border-gray-200 dark:border-gray-700">
          <h2 className="text-2xl font-semibold mb-4 text-gray-900 dark:text-white text-center">
            🛠️ Tech Stack
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {[
              { name: 'React 18', color: 'bg-blue-100 text-blue-800' },
              { name: 'TypeScript', color: 'bg-blue-100 text-blue-800' },
              { name: 'Tailwind CSS', color: 'bg-cyan-100 text-cyan-800' },
              { name: 'Vite', color: 'bg-purple-100 text-purple-800' },
              { name: 'Go Backend', color: 'bg-cyan-100 text-cyan-800' },
              { name: 'Firebase', color: 'bg-orange-100 text-orange-800' },
              { name: 'Nostr Protocol', color: 'bg-green-100 text-green-800' },
              { name: 'GCP Cloud Run', color: 'bg-red-100 text-red-800' }
            ].map((tech, index) => (
              <div key={index} className={`${tech.color} px-3 py-2 rounded-lg text-center text-sm font-medium`}>
                {tech.name}
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <footer className="mt-12 text-center text-gray-500 dark:text-gray-400">
          <p>Built with ❤️ for the music community</p>
        </footer>
      </div>
    </div>
  )
}

export default App