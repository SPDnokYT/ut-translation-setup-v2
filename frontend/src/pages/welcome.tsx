import { Button } from "@/components/ui/button"
import { FaUsers, FaDiscord, FaGithub } from "react-icons/fa"
import { Link } from "wouter"
import { BrowserOpenURL } from "../../wailsjs/runtime"

export default function WelcomePage() {
  return (
    <div className="flex select-none">
      <div>
        <img className="h-dvh max-w-fit" draggable={false} src="banner.webp" />
      </div>

      <div className="mx-10 my-10 flex grow flex-col justify-center">
        <h1 className="text-3xl font-medium tracking-tight text-balance">
          Until Them... <span className="italic">em português! </span>✨
        </h1>

        <p className="mt-5 font-serif text-gray-300">
          Essa tradução foi feita com carinho por fãs, para que mais pessoas
          possam desfrutar de <span className="italic">Until Then</span> em
          nosso belissimo idioma. Esperamos que te emocione tanto quanto nos
          emocionou.
        </p>

        <div className="mt-10">
          <p className="mb-2 flex items-center gap-2 text-lg font-medium text-gray-400">
            <span>
              <FaUsers />
            </span>
            Equipe do Projeto:
          </p>
          <ul className="list-inside list-disc font-serif">
            <li>
              <span className="text-gray-300">(PitterG4)</span> Bernardo
              Hoffmann
            </li>
            <li>
              <span className="text-gray-300">(Percival)</span> Gabriel Araújo
            </li>
            <li>
              <span className="text-gray-300">(Lucasxt)</span> Lucas Silva
            </li>
            <li>
              <span className="text-gray-300">(Yubi)</span> Eduarda Albuquerque
            </li>
            <li>
              <span className="text-gray-300">(Ceci)</span> Cecília
            </li>
            <li>
              <span className="text-gray-300">(flyri0)</span> Francisco
            </li>
          </ul>
        </div>

        <div className="mt-10 w-full self-center">
          <Link to="/pick-file" asChild>
            <Button className="w-full border-2 py-10 text-4xl transition hover:cursor-pointer hover:border-blue-600">
              Vamos lá!
            </Button>
          </Link>
        </div>

        <div className="mt-16 flex w-full justify-center gap-4">
          <Button
            variant="outline"
            onClick={() => BrowserOpenURL("https://discord.gg/MKn6QBVG9g")}
            className="flex h-auto flex-1 flex-col items-center px-6 py-4 hover:cursor-pointer"
          >
            <div className="flex items-center gap-2 text-lg">
              <FaDiscord />
              Discord
            </div>
            <span className="mt-1 text-[10px] font-light tracking-wider uppercase opacity-80">
              Entre na comunidade
            </span>
          </Button>

          <Button
            variant="outline"
            onClick={() =>
              BrowserOpenURL(
                "https://github.com/flyri0/ut-translation-setup-v2"
              )
            }
            className="flex h-auto flex-1 flex-col items-center px-6 py-4 hover:cursor-pointer"
          >
            <div className="flex items-center gap-2 text-lg">
              <FaGithub />
              GitHub
            </div>
            <span className="mt-1 text-[10px] font-light tracking-wider uppercase opacity-80">
              Ver código-fonte
            </span>
          </Button>
        </div>
      </div>
    </div>
  )
}
