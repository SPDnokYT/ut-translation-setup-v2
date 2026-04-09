import { useEffect, useRef, useState } from "react"
import { FaCircleCheck, FaCircleXmark } from "react-icons/fa6"
import { useLocation } from "wouter"

import { Item, ItemContent, ItemMedia, ItemTitle } from "@/components/ui/item"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Spinner } from "@/components/ui/spinner"

import { EventsOn, EventsOff, Quit } from "../../wailsjs/runtime"
import { RunInstallation } from "../../wailsjs/go/main/PckExplorerService"

export default function InstallPage() {
  const MAX_LOGS = 25
  const [, navigate] = useLocation()
  const [step, setStep] = useState("Подготовка к установке...")
  const [logs, setLogs] = useState<string[]>([])
  const [status, setStatus] = useState<"installing" | "success" | "error">(
    "installing"
  )
  const scrollViewportRef = useRef<HTMLDivElement>(null)

  const appendLog = (message: string) => {
    setLogs((prev) => {
      const newLogs = [...prev, message]
      return newLogs.length > MAX_LOGS ? newLogs.slice(-MAX_LOGS) : newLogs
    })
  }

  useEffect(() => {
    RunInstallation().catch((err) => {
      setStatus("error")
      setStep("Критическая ошибка при запуске.")
      appendLog(`[СИСТЕМА] Ошибка при обращении к бэкенду: ${err}`)
    })

    EventsOn("install_step", (currentStep: string) => {
      setStep(currentStep)
    })

    EventsOn("install_log", (msg: string) => {
      appendLog(msg)
    })

    EventsOn("install_error", (err: string) => {
      setStatus("error")
      setStep("Ошибка во время установки.")
      appendLog(`[ОШИБКА] ${err}`)
    })

    EventsOn("install_success", (msg: string) => {
      setStatus("success")
      setStep(msg)
      appendLog("[СИСТЕМА] Процесс успешно завершен!")
    })

    return () => {
      EventsOff("install_step")
      EventsOff("unzip_bin_progress")
      EventsOff("unzip_trans_progress")
      EventsOff("install_log")
      EventsOff("install_error")
      EventsOff("install_success")
    }
  }, [])

  useEffect(() => {
    if (scrollViewportRef.current) {
      scrollViewportRef.current.scrollTop =
        scrollViewportRef.current.scrollHeight
    }
  }, [logs])

  return (
    <div className="mx-10 flex h-screen flex-col items-center justify-center gap-5">
      <Item className="rounded transition-all duration-300" variant="muted">
        <ItemMedia>
          {status === "installing" && <Spinner />}
          {status === "success" && <FaCircleCheck className="text-green-500" />}
          {status === "error" && <FaCircleXmark className="text-red-500" />}
        </ItemMedia>
        <ItemContent>
          <ItemTitle className={status === "error" ? "text-red-500" : ""}>
            {step}
          </ItemTitle>
        </ItemContent>
      </Item>

      <ScrollArea
        ref={scrollViewportRef}
        className="h-1/3 w-full rounded border p-3 font-mono text-xs text-muted-foreground"
      >
        <div className="flex flex-col gap-1">
          {logs.length === 0 ? (
            <span className="italic opacity-50">Ожидание логов...</span>
          ) : (
            logs.map((log, index) => (
              <div
                key={index}
                className={log.startsWith("[ОШИБКА]") ? "text-red-500" : ""}
              >
                {log}
              </div>
            ))
          )}
        </div>
      </ScrollArea>

      {status !== "installing" && (
        <button
          className="mt-4 rounded bg-primary px-6 py-2 text-primary-foreground transition-opacity hover:cursor-pointer"
          onClick={() => {
            if (status === "success") {
              navigate("/finished")
              return
            }

            Quit()
          }}
        >
          {status === "success" ? "Завершить" : "Закрыть установщик"}
        </button>
      )}
    </div>
  )
}
