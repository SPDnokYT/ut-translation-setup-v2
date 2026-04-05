import { Button } from "@/components/ui/button"
import { Quit } from "../../wailsjs/runtime"
import confetti from "canvas-confetti"
import { useEffect } from "react"
import SocialButtons from "@/components/social-buttons"

export default function FinishedPage() {
  useEffect(() => {
    function randomInRange(min: number, max: number): number {
      return Math.random() * (max - min) + min
    }

    const interval = setInterval(() => {
      confetti({
        angle: randomInRange(80, 100),
        colors: ["#FE218B", "#FED700", "#21B0FE", "#6EEB83", "#FF5714"],
        disableForReducedMotion: true,
        drift: randomInRange(-1, 1),
        gravity: randomInRange(0.4, 0.8),
        origin: { x: Math.random(), y: 1.05 },
        particleCount: randomInRange(3, 25),
        scalar: randomInRange(0.4, 1),
        startVelocity: randomInRange(50, 100),
        ticks: 3000,
      })
    }, 50)

    const timeout = setTimeout(() => {
      clearInterval(interval)
    }, 3000)

    return () => {
      clearTimeout(timeout)
      clearInterval(interval)
    }
  }, [])

  return (
    <div className="mx-15 flex h-dvh flex-col items-center justify-center select-none">
      <video
        className="pointer-events-none h-35 rounded"
        draggable={false}
        autoPlay
        loop
        src="cathy.webm"
      />

      <h1 className="mt-2 text-center text-2xl">Instalação Completa! 🎉</h1>

      <p className="text-sm text-muted-foreground sm:mt-5">
        De coração, muito obrigado pela confiança! Foi um esforço de paixão de
        toda a equipe. Caso encontre algum erro ou tenha alguma sugestão, por
        favor, não hesite em nos contar!
        <br />
        No mais, aproveite a viagem. Bom jogo!
      </p>

      <Button className="mt-2 w-full sm:mt-5" size="lg" onClick={Quit}>
        Até Lá!
      </Button>
      <SocialButtons className="mt-2" />
    </div>
  )
}
