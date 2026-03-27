import { Route, Router, Switch } from "wouter"
import { ThemeProvider } from "./components/theme-provider"
import WelcomePage from "./pages/welcome"
import PickFilePage from "./pages/pick-file"
import { TooltipProvider } from "./components/ui/tooltip"

export default function App() {
  return (
    <Router>
      <ThemeProvider defaultTheme="dark">
        <TooltipProvider>
          <Switch>
            <Route path="/" component={WelcomePage} />
            <Route path="/pick-file" component={PickFilePage} />
          </Switch>
        </TooltipProvider>
      </ThemeProvider>
    </Router>
  )
}
