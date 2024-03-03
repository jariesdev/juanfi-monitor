import typer

import sql_app.models
from juanfi_logger import JuanfiLogger
from sql_app.database import SessionLocal, engine
from commands import vendo_add as vendor_adder

app = typer.Typer()


@app.command()
def hello(name: str):
    print(f"Hello {name}")


@app.command()
def goodbye(name: str, formal: bool = False):
    if formal:
        print(f"Goodbye Ms. {name}. Have a good day.")
    else:
        print(f"Bye {name}!")


@app.command()
def vendo_logger():
    db = SessionLocal()
    vendos = db.query(sql_app.models.Vendo).where(sql_app.models.Vendo.is_active == 1).all()
    if (len(vendos) == 0):
        typer.echo("No registered vendo. Please add first.")
        return

    for vendo in vendos:
        typer.echo("Pulling latest from " + vendo.name)
        try:
            logger = JuanfiLogger(vendo)
            logger.run()
        except Exception as e:
            typer.echo(e)


@app.command()
def vendo_add():
    vendor_adder.main()


if __name__ == "__main__":
    app()
