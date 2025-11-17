function App() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border p-4">
        <h1 className="text-2xl font-bold">🔍 ceye</h1>
      </header>
      <main className="container mx-auto p-8">
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-xl font-semibold mb-2">React Dashboard</h2>
          <p className="text-muted-foreground">
            CI/CD Monitoring Dashboard - React + Vite + TypeScript + Tailwind
          </p>
        </div>
      </main>
    </div>
  )
}

export default App
