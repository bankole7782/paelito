# paelito

a book maker and reader

## Why Paelito?
This project would tell if a new version of the book has been published.


## How to Use (Ubuntu)

### Making a book
Paelito currently only supports markdown files.

Take a look at [The Botanum](https://github.com/bankole7782/the_botanum) as a sample to create a book.

Create your project in `$HOME/snap/paelito/common/p` and make it follows the laws in [The Botanum](https://github.com/bankole7782/the_botanum)

Then run `paelito.maker book_folder_name` to create your book.

### Book Details

All the contents of a book folder should not contain any subdirectory.

A book must contain a details.json, toc.txt, cover.png, some markdown files.

Though not necessary a book could have font1.ttf, font2.ttf, book.css and a bg.png

#### details.json
A paelito book starts with the `details.json`.

It must contain a single object with the following fields:
* FullTitle
* Comment
* Author1
* UpdateURL
* BookSourceURL
* Contact
* BookId

If the authors are about three, you need to include the following fields to your details.json: Author1, Author2, Author3

Create a random string of about 20 characters and set it to BookId

The contact should be an email.

#### toc.txt
toc.txt contains the Root Table of Contents of a paelito book.

It contains a toc item seperated by two newlines (\n).

A toc item is the name of a chapter followed by a new line and then the markdown that contains the text of the chapter.

#### On Images
To include an image into your book, please include the image into the book folder.

To display the image, say for example dd.png. You must use the following path for the image `/gba/{BookId}/dd.png` where BookId is the random string in `details.json`.


### Viewing a book
Paelito books must be downloaded and placed in the `$HOME/snap/paelito/common/lib`.

You can now run `paelito`
