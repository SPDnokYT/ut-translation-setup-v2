import { ThemeProvider } from "./components/theme-provider"

export function App() {
  return (
    <ThemeProvider defaultTheme="dark">
      <h1>Hello, Wails!</h1>
    </ThemeProvider>
  )
}

export default App
