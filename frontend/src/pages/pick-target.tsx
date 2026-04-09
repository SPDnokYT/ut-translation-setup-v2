import { useEffect, useState } from "react"
import { FaFileArrowDown, FaFolderOpen, FaFile } from "react-icons/fa6"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Field,
  FieldDescription,
  FieldLabel,
  FieldGroup,
  FieldContent,
  FieldTitle,
} from "@/components/ui/field"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  QuickFind,
  OpenFilePicker,
  CheckFreeSpace,
  SaveSettings,
} from "../../wailsjs/go/main/PickTargetService"
import { toast } from "sonner"
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Toaster } from "@/components/ui/sonner"
import { useLocation } from "wouter"

export default function PickTargetPage() {
  const [, navigate] = useLocation()
  const [targetPath, setTargetPath] = useState<string>("")
  const [isDemo, setIsDemo] = useState<boolean>(false)
  const [isBackupChecked, setIsBackupChecked] = useState<boolean>(false)
  const [isSpaceDialogOpen, setIsSpaceDialogOpen] = useState<boolean>(false)
  const [isValid, setIsValid] = useState<boolean>(false)
  const [isCheckingSpace, setIsCheckingSpace] = useState<boolean>(false)

  async function handleInstallButton() {
    SaveSettings(targetPath, isDemo, isBackupChecked)
    navigate("/install")
  }

  async function handleSuccessfulValidation(
    path: string,
    autoFound: boolean,
    isDemo: boolean
  ) {
    setTargetPath(path)
    setIsDemo(isDemo)

    const hasSpace = await CheckFreeSpace(path)

    if (!hasSpace) {
      setIsSpaceDialogOpen(true)
      setIsValid(false)
    } else {
      setIsValid(true)
      if (autoFound) {
        toast.success("Установка найдена автоматически", {
          description: "Обнаружена установка Until Then в Steam.",
          position: "top-center",
        })
      }
    }
  }

  async function handleManualSelect() {
    try {
      const result = await OpenFilePicker()
      if (result.valid) {
        handleSuccessfulValidation(result.path, false, result.isDemo)
      } else if (result.error !== "not_selected") {
        toast.error("Неверный файл", {
          position: "top-center",
        })
      }
    } catch {
      // TODO: colocar log aqui
    }
  }

  async function handleRetrySpaceCheck() {
    setIsCheckingSpace(true)
    const hasSpace = await CheckFreeSpace(targetPath)
    setIsCheckingSpace(false)

    if (hasSpace) {
      setIsSpaceDialogOpen(false)
      setIsValid(true)
    } else {
      toast.error("По-прежнему недостаточно свободного места.", {
        position: "top-center",
      })
    }
  }

  const handleCancelSpaceCheck = () => {
    setIsSpaceDialogOpen(false)
    setTargetPath("")
    setIsValid(false)
  }

  useEffect(() => {
    const runQuickFind = async () => {
      try {
        const result = await QuickFind()
        if (result.valid) {
          handleSuccessfulValidation(result.path, true, result.isDemo)
        }
      } catch {
        // TODO: put log here
      }
    }

    runQuickFind()
  }, [])

  return (
    <>
      <Toaster />
      <div className="mx-17 flex h-dvh flex-col justify-center select-none">
        <FieldGroup>
          <Field>
            <FieldLabel>Установка игры</FieldLabel>
            <FieldDescription>
              Выберите файл UntilThen.pck в папке с игрой
            </FieldDescription>
            <ButtonGroup>
              <InputGroup>
                <InputGroupAddon
                  align="inline-start"
                  className="pointer-events-none"
                >
                  <FaFile />
                </InputGroupAddon>
                <InputGroupInput
                  disabled
                  value={targetPath}
                  placeholder="Файл не выбран"
                  className="truncate"
                />
              </InputGroup>

              <Button
                className="hover:cursor-pointer"
                variant="outline"
                onClick={handleManualSelect}
              >
                <FaFolderOpen />
                Обзор...
              </Button>
            </ButtonGroup>

            <FieldDescription className="h-5">
              {isValid && (
                <span className="font-medium text-emerald-500">
                  Файл UntilThen.pck успешно проверен.
                </span>
              )}
            </FieldDescription>
          </Field>

          <Field orientation="horizontal">
            <Checkbox
              checked={isBackupChecked}
              onCheckedChange={(value: boolean) => setIsBackupChecked(value)}
            />
            <FieldContent>
              <FieldTitle>Создать резервную копию</FieldTitle>
              <FieldDescription className="w-3/5">
                Сохраняет копию оригинального файла для восстановления стандартного языка (потребуется 2 ГБ дополнительного места).
              </FieldDescription>
            </FieldContent>
          </Field>
        </FieldGroup>

        <Button
          disabled={!isValid}
          onClick={handleInstallButton}
          size="lg"
          className="mt-5 w-full hover:cursor-pointer"
        >
          <FaFileArrowDown />
          Начать установку
        </Button>

        <AlertDialog
          open={isSpaceDialogOpen}
          onOpenChange={setIsSpaceDialogOpen}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Недостаточно места на диске</AlertDialogTitle>
              <AlertDialogDescription>
                На выбранном диске нет необходимых <strong>5 ГБ</strong> свободного места. Освободите место и повторите попытку.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <Button variant="outline" onClick={handleCancelSpaceCheck}>
                Отмена
              </Button>
              <Button
                disabled={isCheckingSpace}
                onClick={handleRetrySpaceCheck}
              >
                {isCheckingSpace ? "Проверка..." : "Повторить попытку"}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </>
  )
}
