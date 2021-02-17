export class RoutingConfig {}

export class BaseWorker {}

export class Worker extends BaseWorker {
  
  start(filePath: string): void {
    console.log(filePath)
  }

  listen(): void {
    console.log('listening')
  }

}
