import { BrowserOpenURL } from "../../wailsjs/runtime"
import { TbWorld } from "react-icons/tb"
import { RiTelegram2Fill } from "react-icons/ri"
import { Button } from "./ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip"
import { cn } from "@/lib/utils"

interface SocialButtonsProps {
  className?: string
}

export default function SocialButtons({ className }: SocialButtonsProps) {
  return (
    <div className={cn("flex w-full gap-1", className)}>
      <Tooltip disableHoverableContent>
        <TooltipTrigger asChild>
          <Button
            className="flex-1 hover:cursor-pointer"
            variant="outline"
            onClick={() => BrowserOpenURL("https://t.me/GameVerbal")}
          >
            <RiTelegram2Fill />
            Telegram
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Присоединиться к сообществу</TooltipContent>
      </Tooltip>

      <Tooltip disableHoverableContent>
        <TooltipTrigger asChild>
          <Button
            className="flex-1 hover:cursor-pointer"
            variant="outline"
            onClick={() =>
              BrowserOpenURL(
                "http://gameverbal.ru/"
              )
            }
          >
            <TbWorld />
            Website
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">Перейти на сайт</TooltipContent>
      </Tooltip>
    </div>
  )
}
