from repository.vendo_repository import VendoRepository


def main():
    name = input('Enter name to add vendo: ')
    url = input('Enter IP address: ')
    api_key = input('Enter API key: ')
    repo = VendoRepository()
    repo.add(name, url, api_key)
    print('added')


if __name__ == '__main__':
    main()
