import { Button } from "@/components/ui/button"
import { FaUsers, FaCode } from "react-icons/fa6"
import { Link } from "wouter"
import ThemeToggle from "@/components/theme-toggle"
import SocialButtons from "@/components/social-buttons"

export default function WelcomePage() {
  const GROUP_NAME = "GameVerbal"
  const INSTALLER_DEV = { alias: "flyri0", name: "Francisco Lyrio" }

  return (
    <div className="flex">
      <img className="h-dvh" draggable={false} src="banner.webp" />

      <div className="mx-5 my-2 self-center">
        <ThemeToggle className="mb-5" />

        <h1 className="mb-1 text-base font-medium tracking-tight text-balance md:text-2xl">
          Until Then... <span className="italic">на русском!</span> ✨
        </h1>

        <p className="font-serif text-xs text-muted-foreground md:text-base">
          Рады видеть тебя здесь!<br />Данный перевод создавался с душой, чтобы как можно больше людей смогли погрузиться в атмосферу <span className="italic">Until Then</span> обходя языковой барьер.
        </p>
		
		<div className="mt-5">
			<p className="flex items-center gap-1 text-sm text-muted-foreground md:text-base">
			  <span>
				<FaUsers />
			  </span>
			  Перевод от:
			</p>
			<p className="font-serif text-base md:text-lg font-medium">
			  {GROUP_NAME}
			</p>
		</div>
		
		<div className="mt-3">
          <p className="flex items-center gap-1 text-sm text-muted-foreground md:text-base">
            <FaCode /> Установщик от:
          </p>
          <p className="font-serif text-xs md:text-sm">
             <span className="text-muted-foreground">({INSTALLER_DEV.alias})</span> {INSTALLER_DEV.name}
          </p>
        </div>

        <Link to="/pick-target" asChild>
          <Button className="mt-5 w-full hover:cursor-pointer md:h-16 md:text-2xl">
            Поехали!
          </Button>
        </Link>

        <SocialButtons className="mt-2" />
      </div>
    </div>
  )
}
